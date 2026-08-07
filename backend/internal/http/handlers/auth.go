package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/config"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	svc      *services.AuthService
	validate *validator.Validate
	cfg      *config.Config
}

func NewAuthHandler(svc *services.AuthService, v *validator.Validate, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, validate: v, cfg: cfg}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	u, err := h.svc.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, dto.UserToResponse(u))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	pair, err := h.svc.Login(r.Context(), req.Email, req.Password, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		respond.Err(w, err)
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	respond.JSON(w, http.StatusOK, dto.AuthResponse{
		AccessToken: pair.AccessToken,
		User:        dto.UserToResponse(pair.User),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil || c.Value == "" {
		respond.Err(w, domain.ErrUnauthorized)
		return
	}
	pair, err := h.svc.Refresh(r.Context(), c.Value, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		if errors.Is(err, domain.ErrTokenReused) || errors.Is(err, domain.ErrTokenRevoked) {
			h.clearRefreshCookie(w)
		}
		respond.Err(w, err)
		return
	}
	h.setRefreshCookie(w, pair.RefreshToken)
	respond.JSON(w, http.StatusOK, dto.AuthResponse{
		AccessToken: pair.AccessToken,
		User:        dto.UserToResponse(pair.User),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(refreshCookieName)
	if c != nil {
		_ = h.svc.Logout(r.Context(), c.Value)
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		respond.Err(w, domain.ErrUnauthorized)
		return
	}
	u, err := h.svc.Me(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.UserToResponse(u))
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		respond.Err(w, domain.ErrUnauthorized)
		return
	}
	var req dto.UpdateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	u, err := h.svc.UpdateProfile(r.Context(), uid, req.DisplayName)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.UserToResponse(u))
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		Domain:   h.cfg.CookieDomain,
		Expires:  time.Now().Add(h.cfg.RefreshTTL),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		Domain:   h.cfg.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
