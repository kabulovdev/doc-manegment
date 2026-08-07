package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
)

type APITokenHandler struct {
	svc      *services.APITokenService
	validate *validator.Validate
}

func NewAPITokenHandler(svc *services.APITokenService, v *validator.Validate) *APITokenHandler {
	return &APITokenHandler{svc: svc, validate: v}
}

func (h *APITokenHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	list, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.APITokenResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.APITokenToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *APITokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.CreateAPITokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	res, err := h.svc.Create(r.Context(), uid, services.CreateAPITokenInput{
		Name:   req.Name,
		Scopes: req.Scopes,
	})
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, dto.CreateAPITokenResponse{
		Token:     dto.APITokenToResponse(res.Token),
		Plaintext: res.Plaintext,
	})
}

func (h *APITokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Revoke(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APITokenHandler) Scopes(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]any{"scopes": services.AllowedAPITokenScopes})
}
