package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileHandler struct {
	svc      *services.FileService
	activity *services.ActivityService
	embed    *services.EmbedService
	validate *validator.Validate
}

func NewFileHandler(svc *services.FileService, activity *services.ActivityService, embed *services.EmbedService, v *validator.Validate) *FileHandler {
	return &FileHandler{svc: svc, activity: activity, embed: embed, validate: v}
}

func (h *FileHandler) record(r *http.Request, uid, sid primitive.ObjectID, action domain.ActivityAction, meta map[string]any) {
	if h.activity == nil {
		return
	}
	h.activity.Record(r.Context(), services.RecordInput{
		UserID:      uid,
		SubjectType: domain.ActivitySubjectFile,
		SubjectID:   sid,
		Action:      action,
		Metadata:    meta,
		IP:          clientIP(r),
		UserAgent:   r.UserAgent(),
	})
}

func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	q := r.URL.Query()
	filter := ports.FileListFilter{
		Query:      q.Get("q"),
		FieldKey:   q.Get("field_key"),
		FieldValue: q.Get("field_value"),
	}
	if s := q.Get("folder_id"); s != "" {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		filter.FolderID = &id
	}
	if s := q.Get("storage_id"); s != "" {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		filter.StorageID = &id
	}
	if s := q.Get("tag_id"); s != "" {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		filter.TagID = &id
	}
	if s := q.Get("starred"); s == "true" || s == "1" {
		v := true
		filter.Starred = &v
	} else if s == "false" || s == "0" {
		v := false
		filter.Starred = &v
	}
	if s := q.Get("status"); s != "" {
		st := domain.FileStatus(s)
		switch st {
		case domain.FileStatusReady, domain.FileStatusUploading, domain.FileStatusDeleted:
			filter.Status = &st
		}
	}
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			filter.Limit = n
		}
	}
	if s := q.Get("skip"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			filter.Skip = n
		}
	}
	list, err := h.svc.List(r.Context(), uid, filter)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.FileResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FileToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *FileHandler) InitUpload(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.InitUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	storageID, err := primitive.ObjectIDFromHex(req.StorageID)
	if err != nil {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	var folderID *primitive.ObjectID
	if req.FolderID != nil && *req.FolderID != "" {
		id, err := primitive.ObjectIDFromHex(*req.FolderID)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		folderID = &id
	}
	tagIDs := make([]primitive.ObjectID, 0, len(req.TagIDs))
	for _, s := range req.TagIDs {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		tagIDs = append(tagIDs, id)
	}
	fields := make([]domain.CustomField, 0, len(req.CustomFields))
	for _, cf := range req.CustomFields {
		fields = append(fields, domain.CustomField{Key: cf.Key, Value: cf.Value, Type: domain.CustomFieldType(cf.Type)})
	}
	res, err := h.svc.InitUpload(r.Context(), uid, services.InitUploadInput{
		StorageID:    storageID,
		FolderID:     folderID,
		Name:         req.Name,
		Size:         req.Size,
		Mime:         req.Mime,
		CustomFields: fields,
		TagIDs:       tagIDs,
	})
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, dto.InitUploadResponse{
		FileID:    res.FileID.Hex(),
		UploadID:  res.UploadID,
		ObjectKey: res.ObjectKey,
		PartSize:  res.PartSize,
		Multipart: res.Multipart,
	})
}

func (h *FileHandler) UploadPart(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	partStr := r.URL.Query().Get("part_number")
	partNum, err := strconv.Atoi(partStr)
	if err != nil {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	if r.ContentLength <= 0 {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	etag, err := h.svc.UploadPart(r.Context(), uid, fileID, partNum, r.Body, r.ContentLength)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"etag": etag, "part_number": partNum})
}

func (h *FileHandler) Complete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.CompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	parts := make([]ports.Part, 0, len(req.Parts))
	for _, p := range req.Parts {
		parts = append(parts, ports.Part{PartNumber: p.PartNumber, ETag: p.ETag})
	}
	f, err := h.svc.CompleteUpload(r.Context(), uid, fileID, services.CompleteInput{Parts: parts})
	if err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, f.ID, "upload_complete", map[string]any{
		"name":       f.Name,
		"size_bytes": f.SizeBytes,
		"mime":       f.MimeType,
	})
	if h.embed != nil {
		h.embed.IndexFileAsync(uid, f.ID)
	}
	respond.JSON(w, http.StatusOK, dto.FileToResponse(f))
}

func (h *FileHandler) Abort(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.AbortUpload(r.Context(), uid, fileID); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FileHandler) UploadSingle(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	if r.ContentLength <= 0 {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	mime := r.Header.Get("Content-Type")
	if strings.HasPrefix(mime, "multipart/") {
		respond.Err(w, fmt.Errorf("%w: raw body expected", domain.ErrInvalidInput))
		return
	}
	f, err := h.svc.UploadSingle(r.Context(), uid, fileID, r.Body, r.ContentLength, mime)
	if err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, f.ID, "upload_complete", map[string]any{
		"name":       f.Name,
		"size_bytes": f.SizeBytes,
		"mime":       f.MimeType,
	})
	if h.embed != nil {
		h.embed.IndexFileAsync(uid, f.ID)
	}
	respond.JSON(w, http.StatusOK, dto.FileToResponse(f))
}

func (h *FileHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.UpdateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	in := services.UpdateInput{Name: req.Name}
	if req.TagIDs != nil {
		ids := make([]primitive.ObjectID, 0, len(*req.TagIDs))
		for _, s := range *req.TagIDs {
			tid, err := primitive.ObjectIDFromHex(s)
			if err != nil {
				respond.Err(w, domain.ErrInvalidInput)
				return
			}
			ids = append(ids, tid)
		}
		in.TagIDs = &ids
	}
	if req.CustomFields != nil {
		fields := make([]domain.CustomField, 0, len(*req.CustomFields))
		for _, cf := range *req.CustomFields {
			fields = append(fields, domain.CustomField{Key: cf.Key, Value: cf.Value, Type: domain.CustomFieldType(cf.Type)})
		}
		in.CustomFields = &fields
	}
	f, err := h.svc.Update(r.Context(), uid, id, in)
	if err != nil {
		respond.Err(w, err)
		return
	}
	if req.Name != nil {
		h.record(r, uid, f.ID, "rename", map[string]any{"name": f.Name})
	} else {
		h.record(r, uid, f.ID, "update", nil)
	}
	respond.JSON(w, http.StatusOK, dto.FileToResponse(f))
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	f, err := h.svc.Find(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.FileToResponse(f))
}

func (h *FileHandler) Star(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.StarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	f, err := h.svc.SetStarred(r.Context(), uid, id, req.Starred)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.FileToResponse(f))
}

func (h *FileHandler) Restore(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	f, err := h.svc.Restore(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, f.ID, "restore", nil)
	if h.embed != nil {
		h.embed.IndexFileAsync(uid, f.ID)
	}
	respond.JSON(w, http.StatusOK, dto.FileToResponse(f))
}

func (h *FileHandler) Purge(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.PermanentlyDelete(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, id, "purge", nil)
	if h.embed != nil {
		h.embed.RemoveFile(r.Context(), uid, id)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, id, "delete", nil)
	if h.embed != nil {
		h.embed.RemoveFile(r.Context(), uid, id)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FileHandler) Stream(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	rng := parseRange(r.Header.Get("Range"))
	rc, f, info, err := h.svc.Stream(r.Context(), uid, id, rng)
	if err != nil {
		respond.Err(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", info.ContentType)
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", f.MimeType)
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", `inline; filename="`+sanitizeFilename(f.Name)+`"`)
	if info.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	}
	status := http.StatusOK
	if rng != nil {
		status = http.StatusPartialContent
	}
	w.WriteHeader(status)
	if _, err := io.Copy(w, rc); err != nil && !errors.Is(err, io.EOF) {
		return
	}
}

func parseRange(h string) *ports.Range {
	if h == "" {
		return nil
	}
	if !strings.HasPrefix(h, "bytes=") {
		return nil
	}
	parts := strings.SplitN(strings.TrimPrefix(h, "bytes="), "-", 2)
	if len(parts) != 2 {
		return nil
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil
	}
	end := int64(-1)
	if parts[1] != "" {
		e, err := strconv.ParseInt(parts[1], 10, 64)
		if err == nil {
			end = e
		}
	}
	return &ports.Range{Start: start, End: end}
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, "\r", "")
	return name
}

// stop unused import warnings in handlers package when file size is 0
var _ = chi.URLParam
