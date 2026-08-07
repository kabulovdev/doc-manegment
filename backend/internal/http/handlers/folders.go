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

type FolderHandler struct {
	svc      *services.FolderService
	activity *services.ActivityService
	validate *validator.Validate
}

func NewFolderHandler(svc *services.FolderService, activity *services.ActivityService, v *validator.Validate) *FolderHandler {
	return &FolderHandler{svc: svc, activity: activity, validate: v}
}

func (h *FolderHandler) record(r *http.Request, uid, sid primitive.ObjectID, action domain.ActivityAction, meta map[string]any) {
	if h.activity == nil {
		return
	}
	h.activity.Record(r.Context(), services.RecordInput{
		UserID:      uid,
		SubjectType: domain.ActivitySubjectFolder,
		SubjectID:   sid,
		Action:      action,
		Metadata:    meta,
		IP:          clientIP(r),
		UserAgent:   r.UserAgent(),
	})
}

func (h *FolderHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var parentID *primitive.ObjectID
	if s := r.URL.Query().Get("parent_id"); s != "" {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		parentID = &id
	}
	children, err := h.svc.List(r.Context(), uid, parentID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.FolderResponse, 0, len(children))
	for i := range children {
		out = append(out, dto.FolderToResponse(&children[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *FolderHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	breadcrumbs, _ := h.svc.Breadcrumbs(r.Context(), uid, id)
	children, _ := h.svc.List(r.Context(), uid, &id)
	bcOut := make([]dto.FolderResponse, 0, len(breadcrumbs))
	for i := range breadcrumbs {
		bcOut = append(bcOut, dto.FolderToResponse(&breadcrumbs[i]))
	}
	childOut := make([]dto.FolderResponse, 0, len(children))
	for i := range children {
		childOut = append(childOut, dto.FolderToResponse(&children[i]))
	}
	resp := dto.FolderToResponse(f)
	respond.JSON(w, http.StatusOK, dto.FolderListResponse{
		Folder:      &resp,
		Breadcrumbs: bcOut,
		Children:    childOut,
	})
}

func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.CreateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	var parentID *primitive.ObjectID
	if req.ParentID != nil && *req.ParentID != "" {
		id, err := primitive.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		parentID = &id
	}
	f, err := h.svc.Create(r.Context(), uid, req.Name, parentID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, f.ID, "create", map[string]any{"name": f.Name})
	respond.JSON(w, http.StatusCreated, dto.FolderToResponse(f))
}

func (h *FolderHandler) Rename(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.RenameFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.svc.Rename(r.Context(), uid, id, req.Name); err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, id, "rename", map[string]any{"name": req.Name})
	w.WriteHeader(http.StatusNoContent)
}

func (h *FolderHandler) Move(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.MoveFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	var newParent *primitive.ObjectID
	if req.NewParentID != nil && *req.NewParentID != "" {
		pid, err := primitive.ObjectIDFromHex(*req.NewParentID)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		newParent = &pid
	}
	if err := h.svc.Move(r.Context(), uid, id, newParent); err != nil {
		respond.Err(w, err)
		return
	}
	meta := map[string]any{}
	if newParent != nil {
		meta["new_parent_id"] = newParent.Hex()
	} else {
		meta["new_parent_id"] = nil
	}
	h.record(r, uid, id, "move", meta)
	w.WriteHeader(http.StatusNoContent)
}

func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	recursive := r.URL.Query().Get("recursive") == "true"
	if err := h.svc.Delete(r.Context(), uid, id, recursive); err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, id, "delete", map[string]any{"recursive": recursive})
	w.WriteHeader(http.StatusNoContent)
}
