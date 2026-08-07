package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CreateAPITokenRequest struct {
	Name   string   `json:"name" validate:"required,min=1,max=100"`
	Scopes []string `json:"scopes"`
}

type APITokenResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type CreateAPITokenResponse struct {
	Token     APITokenResponse `json:"token"`
	Plaintext string           `json:"plaintext"`
}

func APITokenToResponse(t *domain.APIToken) APITokenResponse {
	scopes := t.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	return APITokenResponse{
		ID:         t.ID.Hex(),
		Name:       t.Name,
		Prefix:     t.Prefix,
		Scopes:     scopes,
		CreatedAt:  t.CreatedAt,
		LastUsedAt: t.LastUsedAt,
		RevokedAt:  t.RevokedAt,
	}
}
