package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AllowedAPITokenScopes lists the scopes clients may request. They're stored
// on the token for display and future fine-grained enforcement; right now any
// valid (non-revoked) token grants the same access as a JWT session.
var AllowedAPITokenScopes = []string{
	"files:read",
	"files:write",
	"folders:read",
	"folders:write",
	"tags:read",
	"tags:write",
	"storages:read",
	"storages:write",
	"shares:read",
	"shares:write",
	"activity:read",
}

type APITokenService struct {
	repo ports.APITokenRepo
}

func NewAPITokenService(repo ports.APITokenRepo) *APITokenService {
	return &APITokenService{repo: repo}
}

type CreateAPITokenInput struct {
	Name   string
	Scopes []string
}

type CreateAPITokenResult struct {
	Token     *domain.APIToken
	Plaintext string
}

func (s *APITokenService) Create(ctx context.Context, userID primitive.ObjectID, in CreateAPITokenInput) (*CreateAPITokenResult, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 100 {
		return nil, domain.ErrInvalidInput
	}
	scopes, err := normalizeScopes(in.Scopes)
	if err != nil {
		return nil, err
	}
	plaintext, hash, prefix, err := auth.NewAPIToken()
	if err != nil {
		return nil, err
	}
	t := &domain.APIToken{
		UserID: userID,
		Name:   name,
		Hash:   hash,
		Prefix: prefix,
		Scopes: scopes,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return &CreateAPITokenResult{Token: t, Plaintext: plaintext}, nil
}

func (s *APITokenService) List(ctx context.Context, userID primitive.ObjectID) ([]domain.APIToken, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *APITokenService) Revoke(ctx context.Context, userID, id primitive.ObjectID) error {
	return s.repo.Revoke(ctx, userID, id)
}

// VerifyBearer checks a plaintext `doc_*` token; returns the owning user if
// the token is valid and not revoked. The last-used timestamp is bumped
// best-effort.
func (s *APITokenService) VerifyBearer(ctx context.Context, plaintext string) (*domain.APIToken, error) {
	if !strings.HasPrefix(plaintext, domain.APITokenPrefix) {
		return nil, domain.ErrUnauthorized
	}
	hash := auth.HashAPIToken(plaintext)
	t, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if t.RevokedAt != nil {
		return nil, domain.ErrUnauthorized
	}
	_ = s.repo.TouchLastUsed(ctx, t.ID, time.Now().UTC())
	return t, nil
}

func normalizeScopes(in []string) ([]string, error) {
	if len(in) == 0 {
		out := make([]string, len(AllowedAPITokenScopes))
		copy(out, AllowedAPITokenScopes)
		return out, nil
	}
	allowed := make(map[string]struct{}, len(AllowedAPITokenScopes))
	for _, s := range AllowedAPITokenScopes {
		allowed[s] = struct{}{}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if _, ok := allowed[s]; !ok {
			return nil, fmt.Errorf("%w: unknown scope %q", domain.ErrInvalidInput, s)
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: at least one scope required", domain.ErrInvalidInput)
	}
	return out, nil
}
