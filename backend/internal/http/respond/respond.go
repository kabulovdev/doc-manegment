package respond

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type errorBody struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("respond encode", "err", err)
	}
}

func Err(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		JSON(w, http.StatusNotFound, errorBody{Error: "not_found", Message: err.Error()})
	case errors.Is(err, domain.ErrConflict):
		JSON(w, http.StatusConflict, errorBody{Error: "conflict", Message: err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		JSON(w, http.StatusBadRequest, errorBody{Error: "invalid_input", Message: err.Error()})
	case errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrInvalidCreds),
		errors.Is(err, domain.ErrTokenExpired),
		errors.Is(err, domain.ErrTokenRevoked),
		errors.Is(err, domain.ErrTokenReused):
		JSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized", Message: err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		JSON(w, http.StatusForbidden, errorBody{Error: "forbidden", Message: err.Error()})
	default:
		slog.Error("internal error", "err", err)
		JSON(w, http.StatusInternalServerError, errorBody{Error: "internal", Message: "internal server error"})
	}
}

func ValidationErr(w http.ResponseWriter, err error) {
	JSON(w, http.StatusBadRequest, errorBody{Error: "validation_error", Message: err.Error()})
}
