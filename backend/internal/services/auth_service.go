package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct {
	users      ports.UserRepo
	refreshes  ports.RefreshTokenRepo
	jwt        *auth.JWTIssuer
	refreshTTL time.Duration
}

func NewAuthService(users ports.UserRepo, refreshes ports.RefreshTokenRepo, jwt *auth.JWTIssuer, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, refreshes: refreshes, jwt: jwt, refreshTTL: refreshTTL}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(password) < 8 {
		return nil, domain.ErrInvalidInput
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		Email:        email,
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(displayName),
	}
	if u.DisplayName == "" {
		u.DisplayName = strings.Split(email, "@")[0]
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, ua, ip string) (*TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCreds
		}
		return nil, err
	}
	ok, err := auth.VerifyPassword(password, u.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidCreds
	}
	_ = s.users.TouchLastLogin(ctx, u.ID)
	return s.issueTokens(ctx, u, primitive.NilObjectID, ua, ip)
}

func (s *AuthService) Me(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *AuthService) UpdateProfile(ctx context.Context, id primitive.ObjectID, displayName string) (*domain.User, error) {
	name := strings.TrimSpace(displayName)
	if name == "" || len(name) > 100 {
		return nil, domain.ErrInvalidInput
	}
	if err := s.users.UpdateDisplayName(ctx, id, name); err != nil {
		return nil, err
	}
	return s.users.FindByID(ctx, id)
}

func (s *AuthService) Refresh(ctx context.Context, refreshPlain, ua, ip string) (*TokenPair, error) {
	hash := auth.HashRefreshToken(refreshPlain)
	tok, err := s.refreshes.FindByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if tok.Revoked {
		_ = s.refreshes.RevokeFamily(ctx, tok.FamilyID)
		return nil, domain.ErrTokenRevoked
	}
	if time.Now().After(tok.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}
	if tok.ReplacedBy != nil {
		_ = s.refreshes.RevokeFamily(ctx, tok.FamilyID)
		return nil, domain.ErrTokenReused
	}
	u, err := s.users.FindByID(ctx, tok.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	pair, newTok, err := s.issueTokensWithID(ctx, u, tok.FamilyID, ua, ip)
	if err != nil {
		return nil, err
	}
	if err := s.refreshes.MarkReplaced(ctx, tok.ID, newTok.ID); err != nil {
		return nil, err
	}
	return pair, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshPlain string) error {
	if refreshPlain == "" {
		return nil
	}
	hash := auth.HashRefreshToken(refreshPlain)
	tok, err := s.refreshes.FindByHash(ctx, hash)
	if err != nil {
		return nil
	}
	return s.refreshes.RevokeFamily(ctx, tok.FamilyID)
}

func (s *AuthService) issueTokens(ctx context.Context, u *domain.User, existingFamily primitive.ObjectID, ua, ip string) (*TokenPair, error) {
	pair, _, err := s.issueTokensWithID(ctx, u, existingFamily, ua, ip)
	return pair, err
}

func (s *AuthService) issueTokensWithID(ctx context.Context, u *domain.User, existingFamily primitive.ObjectID, ua, ip string) (*TokenPair, *domain.RefreshToken, error) {
	access, err := s.jwt.IssueAccess(u.ID)
	if err != nil {
		return nil, nil, err
	}
	plain, hash, err := auth.NewRefreshToken()
	if err != nil {
		return nil, nil, err
	}
	family := existingFamily
	if family.IsZero() {
		family = primitive.NewObjectID()
	}
	rt := &domain.RefreshToken{
		UserID:    u.ID,
		TokenHash: hash,
		FamilyID:  family,
		ExpiresAt: time.Now().Add(s.refreshTTL).UTC(),
		CreatedAt: time.Now().UTC(),
		UserAgent: ua,
		IP:        ip,
	}
	if err := s.refreshes.Create(ctx, rt); err != nil {
		return nil, nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: plain, User: u}, rt, nil
}
