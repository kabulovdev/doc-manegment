package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageHandler struct {
	svc      *services.StorageService
	validate *validator.Validate
}

func NewStorageHandler(svc *services.StorageService, v *validator.Validate) *StorageHandler {
	return &StorageHandler{svc: svc, validate: v}
}

func (h *StorageHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	list, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.StorageResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.StorageToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *StorageHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.CreateStorageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	sc, err := h.svc.Create(r.Context(), uid, services.CreateStorageInput{
		DisplayName:    req.DisplayName,
		Provider:       domain.StorageProvider(req.Provider),
		Endpoint:       req.Endpoint,
		Region:         req.Region,
		Bucket:         req.Bucket,
		AccessKey:      req.AccessKey,
		SecretKey:      req.SecretKey,
		ForcePathStyle: req.ForcePathStyle,
	})
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, dto.StorageToResponse(sc))
}

func (h *StorageHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	sc, err := h.svc.Find(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.StorageToResponse(sc))
}

func (h *StorageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	force := r.URL.Query().Get("force") == "true"
	if err := h.svc.Delete(r.Context(), uid, id, force); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StorageHandler) Test(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.TestConnection(r.Context(), uid, id); err != nil {
		respond.JSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *StorageHandler) Resync(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Resync(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	sc, err := h.svc.Find(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.StorageToResponse(sc))
}

func objectID(w http.ResponseWriter, r *http.Request) (primitive.ObjectID, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		respond.Err(w, domain.ErrInvalidInput)
		return primitive.NilObjectID, false
	}
	return id, true
}

// clientIP returns the best-effort originating IP, preferring the remote addr
// which go-chi's RealIP middleware populates from proxies.
func clientIP(r *http.Request) string {
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}
