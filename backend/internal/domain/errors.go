package domain

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrForbidden        = errors.New("forbidden")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidInput     = errors.New("invalid input")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenRevoked     = errors.New("token revoked")
	ErrTokenReused      = errors.New("token reused")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidCreds     = errors.New("invalid credentials")
	ErrNoAIConfig       = errors.New("no AI provider configured for this capability")
)
