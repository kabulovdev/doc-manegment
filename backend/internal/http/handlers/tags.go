package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TagHandler struct {
	svc      *services.TagService
	validate *validator.Validate
}

func NewTagHandler(svc *services.TagService, v *validator.Validate) *TagHandler {
	return &TagHandler{svc: svc, validate: v}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	tags, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.TagResponse, 0, len(tags))
	for i := range tags {
		out = append(out, dto.TagToResponse(&tags[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	t, err := h.svc.Create(r.Context(), uid, req.Name, req.Color)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, dto.TagToResponse(t))
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.svc.Update(r.Context(), uid, id, req.Name, req.Color); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) Attach(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	tagID, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.AttachTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	targetID, err := primitive.ObjectIDFromHex(req.TargetID)
	if err != nil {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.Attach(r.Context(), uid, tagID, domain.TagTargetType(req.TargetType), targetID); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) Detach(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	tagID, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.AttachTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	targetID, err := primitive.ObjectIDFromHex(req.TargetID)
	if err != nil {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	if err := h.svc.Detach(r.Context(), uid, tagID, domain.TagTargetType(req.TargetType), targetID); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
