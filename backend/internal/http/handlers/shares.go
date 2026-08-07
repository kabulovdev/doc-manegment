package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShareHandler struct {
	svc      *services.ShareService
	activity *services.ActivityService
	validate *validator.Validate
}

func NewShareHandler(svc *services.ShareService, activity *services.ActivityService, v *validator.Validate) *ShareHandler {
	return &ShareHandler{svc: svc, activity: activity, validate: v}
}

func (h *ShareHandler) record(r *http.Request, uid, sid primitive.ObjectID, action domain.ActivityAction, meta map[string]any) {
	if h.activity == nil {
		return
	}
	h.activity.Record(r.Context(), services.RecordInput{
		UserID:      uid,
		SubjectType: domain.ActivitySubjectShare,
		SubjectID:   sid,
		Action:      action,
		Metadata:    meta,
		IP:          clientIP(r),
		UserAgent:   r.UserAgent(),
	})
}

func (h *ShareHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	list, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.ShareResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.ShareToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *ShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.CreateShareRequest
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
	link, err := h.svc.Create(r.Context(), uid, services.CreateShareInput{
		TargetType: domain.ShareTargetType(req.TargetType),
		TargetID:   targetID,
		ExpiresAt:  req.ExpiresAt,
		Password:   req.Password,
		OneTimeUse: req.OneTimeUse,
	})
	if err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, link.ID, "create", map[string]any{
		"target_type":  string(link.TargetType),
		"target_id":    link.TargetID.Hex(),
		"has_password": link.PasswordHash != "",
		"one_time_use": link.OneTimeUse,
	})
	respond.JSON(w, http.StatusCreated, dto.ShareToResponse(link))
}

func (h *ShareHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	link, err := h.svc.Find(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	logs, _ := h.svc.AccessLog(r.Context(), uid, id, 50)
	entries := make([]dto.ShareAccessLogEntry, 0, len(logs))
	for _, l := range logs {
		entries = append(entries, dto.ShareAccessLogEntry{
			AccessedAt: l.AccessedAt, IP: l.IP, UserAgent: l.UserAgent, BytesStreamed: l.BytesStreamed,
		})
	}
	respond.JSON(w, http.StatusOK, map[string]any{
		"share":  dto.ShareToResponse(link),
		"access": entries,
	})
}

func (h *ShareHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Revoke(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	h.record(r, uid, id, "revoke", nil)
	w.WriteHeader(http.StatusNoContent)
}

func shareStatusError(err error) int {
	switch {
	case errors.Is(err, services.ErrShareExpired),
		errors.Is(err, services.ErrShareConsumed),
		errors.Is(err, services.ErrShareRevoked):
		return http.StatusGone
	case errors.Is(err, services.ErrPasswordNeeded),
		errors.Is(err, services.ErrPasswordWrong):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}
