package handlers

import (
	"net/http"
	"strconv"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityHandler struct {
	svc *services.ActivityService
}

func NewActivityHandler(svc *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func parseLimit(q string, def int64) int64 {
	if q == "" {
		return def
	}
	if n, err := strconv.ParseInt(q, 10, 64); err == nil && n > 0 && n <= 200 {
		return n
	}
	return def
}

func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	q := r.URL.Query()
	subjectIDHex := q.Get("subject_id")
	limit := parseLimit(q.Get("limit"), 50)

	if subjectIDHex == "" {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	sid, err := primitive.ObjectIDFromHex(subjectIDHex)
	if err != nil {
		respond.Err(w, domain.ErrInvalidInput)
		return
	}
	list, err := h.svc.ListSubject(r.Context(), uid, sid, limit)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.ActivityResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.ActivityToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *ActivityHandler) Recent(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	limit := parseLimit(r.URL.Query().Get("limit"), 20)
	list, err := h.svc.ListRecent(r.Context(), uid, limit)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.ActivityResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.ActivityToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}
