package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/config"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicShareHandler struct {
	svc      *services.ShareService
	files    services.FileFinder
	folders  services.FolderFinder
	validate *validator.Validate
	cfg      *config.Config
}

func NewPublicShareHandler(svc *services.ShareService, files services.FileFinder, folders services.FolderFinder, v *validator.Validate, cfg *config.Config) *PublicShareHandler {
	return &PublicShareHandler{svc: svc, files: files, folders: folders, validate: v, cfg: cfg}
}

const shareSessionCookie = "share_session"

func (h *PublicShareHandler) Metadata(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	link, err := h.svc.Resolve(r.Context(), token)
	if err != nil {
		h.writeShareErr(w, err)
		return
	}
	unlocked := link.PasswordHash == "" || h.hasValidSessionCookie(r, token)
	resp := dto.PublicShareMetadata{
		TargetType:       string(link.TargetType),
		RequiresPassword: link.PasswordHash != "",
		Unlocked:         unlocked,
	}
	// If no password required or unlocked, include richer metadata.
	switch link.TargetType {
	case domain.ShareTargetFile:
		f, err := h.files.Find(r.Context(), link.UserID, link.TargetID)
		if err == nil {
			resp.Name = f.Name
			if unlocked {
				resp.MimeType = f.MimeType
				resp.SizeBytes = f.SizeBytes
			}
		}
	case domain.ShareTargetFolder:
		fld, err := h.folders.Find(r.Context(), link.UserID, link.TargetID)
		if err == nil {
			resp.Name = fld.Name
			if unlocked {
				resp.FolderID = fld.ID.Hex()
			}
		}
	}
	respond.JSON(w, http.StatusOK, resp)
}

func (h *PublicShareHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	link, err := h.svc.Resolve(r.Context(), token)
	if err != nil {
		h.writeShareErr(w, err)
		return
	}
	var req dto.UnlockShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.svc.VerifyPassword(link, req.Password); err != nil {
		h.writeShareErr(w, err)
		return
	}
	h.setSessionCookie(w, token)
	respond.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *PublicShareHandler) Content(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	link, err := h.svc.Resolve(r.Context(), token)
	if err != nil {
		h.writeShareErr(w, err)
		return
	}
	if link.PasswordHash != "" && !h.hasValidSessionCookie(r, token) {
		h.writeShareErr(w, services.ErrPasswordNeeded)
		return
	}

	var rc io.ReadCloser
	var f *domain.File
	var info struct {
		size     int64
		mime     string
		status   int
	}
	info.status = http.StatusOK
	rng := parseRange(r.Header.Get("Range"))
	if rng != nil {
		info.status = http.StatusPartialContent
	}

	switch link.TargetType {
	case domain.ShareTargetFile:
		rcx, fx, objInfo, err := h.svc.OpenFile(r.Context(), link, rng)
		if err != nil {
			h.writeShareErr(w, err)
			return
		}
		rc = rcx
		f = fx
		info.size = objInfo.Size
		info.mime = objInfo.ContentType
	case domain.ShareTargetFolder:
		// For folder shares, the client must pass ?file_id= pointing to a file inside the subtree.
		fidStr := r.URL.Query().Get("file_id")
		fid, err := primitive.ObjectIDFromHex(fidStr)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		rcx, fx, objInfo, err := h.svc.OpenFolderFile(r.Context(), link, fid, rng)
		if err != nil {
			h.writeShareErr(w, err)
			return
		}
		rc = rcx
		f = fx
		info.size = objInfo.Size
		info.mime = objInfo.ContentType
	default:
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	defer rc.Close()

	mime := info.mime
	if mime == "" {
		mime = f.MimeType
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", `inline; filename="`+sanitizeFilename(f.Name)+`"`)
	if info.size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(info.size, 10))
	}
	w.WriteHeader(info.status)

	n, err := io.Copy(w, rc)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	h.svc.LogAccess(r.Context(), link.ID, ip, r.UserAgent(), n)
	if link.OneTimeUse && link.ConsumedAt == nil {
		_ = h.svc.ConsumeIfOneTime(r.Context(), link)
	}
}

func (h *PublicShareHandler) FolderChildren(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	link, err := h.svc.Resolve(r.Context(), token)
	if err != nil {
		h.writeShareErr(w, err)
		return
	}
	if link.PasswordHash != "" && !h.hasValidSessionCookie(r, token) {
		h.writeShareErr(w, services.ErrPasswordNeeded)
		return
	}
	if link.TargetType != domain.ShareTargetFolder {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	folderID := link.TargetID
	if s := r.URL.Query().Get("folder_id"); s != "" {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		folderID = id
	}
	folders, files, err := h.svc.ListFolder(r.Context(), link, folderID)
	if err != nil {
		h.writeShareErr(w, err)
		return
	}
	cur, _ := h.folders.Find(r.Context(), link.UserID, folderID)
	resp := dto.PublicFolderListing{
		Children: make([]dto.PublicFolderEntry, 0, len(folders)),
		Files:    make([]dto.PublicFileEntry, 0, len(files)),
	}
	if cur != nil {
		resp.Folder = &dto.PublicFolderEntry{ID: cur.ID.Hex(), Name: cur.Name}
	}
	for _, ch := range folders {
		resp.Children = append(resp.Children, dto.PublicFolderEntry{ID: ch.ID.Hex(), Name: ch.Name})
	}
	for _, f := range files {
		resp.Files = append(resp.Files, dto.PublicFileEntry{
			ID: f.ID.Hex(), Name: f.Name, MimeType: f.MimeType, SizeBytes: f.SizeBytes,
		})
	}
	respond.JSON(w, http.StatusOK, resp)
}

func (h *PublicShareHandler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     shareSessionCookie + "_" + firstN(token, 16),
		Value:    "1",
		Path:     "/api/v1/share/" + token,
		Expires:  time.Now().Add(2 * time.Hour),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *PublicShareHandler) hasValidSessionCookie(r *http.Request, token string) bool {
	c, err := r.Cookie(shareSessionCookie + "_" + firstN(token, 16))
	return err == nil && c.Value == "1"
}

func (h *PublicShareHandler) writeShareErr(w http.ResponseWriter, err error) {
	status := shareStatusError(err)
	msg := err.Error()
	respond.JSON(w, status, map[string]string{"error": "share_error", "message": msg})
}

func firstN(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// Prevent chi import from being flagged as unused when file has no direct chi calls.
var _ = strings.TrimSpace
