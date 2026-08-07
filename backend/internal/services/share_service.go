package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileFinder interface {
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.File, error)
}

type FolderFinder interface {
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.Folder, error)
}

type ShareService struct {
	shares   ports.ShareRepo
	logs     ports.ShareAccessLogRepo
	files    ports.FileRepo
	folders  ports.FolderRepo
	storage  *StorageService
}

func NewShareService(shares ports.ShareRepo, logs ports.ShareAccessLogRepo, files ports.FileRepo, folders ports.FolderRepo, storage *StorageService) *ShareService {
	return &ShareService{shares: shares, logs: logs, files: files, folders: folders, storage: storage}
}

type CreateShareInput struct {
	TargetType domain.ShareTargetType
	TargetID   primitive.ObjectID
	ExpiresAt  *time.Time
	Password   string
	OneTimeUse bool
}

func (s *ShareService) Create(ctx context.Context, userID primitive.ObjectID, in CreateShareInput) (*domain.ShareLink, error) {
	switch in.TargetType {
	case domain.ShareTargetFile:
		f, err := s.files.Find(ctx, userID, in.TargetID)
		if err != nil {
			return nil, err
		}
		if f.Status != domain.FileStatusReady {
			return nil, fmt.Errorf("%w: file is not ready", domain.ErrInvalidInput)
		}
	case domain.ShareTargetFolder:
		if _, err := s.folders.Find(ctx, userID, in.TargetID); err != nil {
			return nil, err
		}
	default:
		return nil, domain.ErrInvalidInput
	}

	token, err := generateShareToken()
	if err != nil {
		return nil, err
	}

	link := &domain.ShareLink{
		UserID:     userID,
		Token:      token,
		TargetType: in.TargetType,
		TargetID:   in.TargetID,
		ExpiresAt:  in.ExpiresAt,
		OneTimeUse: in.OneTimeUse,
	}
	if in.Password != "" {
		hash, err := auth.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}
		link.PasswordHash = hash
	}
	if err := s.shares.Create(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *ShareService) List(ctx context.Context, userID primitive.ObjectID) ([]domain.ShareLink, error) {
	return s.shares.ListByUser(ctx, userID)
}

func (s *ShareService) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.ShareLink, error) {
	return s.shares.Find(ctx, userID, id)
}

func (s *ShareService) Revoke(ctx context.Context, userID, id primitive.ObjectID) error {
	return s.shares.Revoke(ctx, userID, id)
}

func (s *ShareService) AccessLog(ctx context.Context, userID, shareID primitive.ObjectID, limit int64) ([]domain.ShareAccessLog, error) {
	if _, err := s.shares.Find(ctx, userID, shareID); err != nil {
		return nil, err
	}
	return s.logs.TailByShare(ctx, shareID, limit)
}

var (
	ErrShareExpired  = errors.New("share expired")
	ErrShareRevoked  = errors.New("share revoked")
	ErrShareConsumed = errors.New("share consumed")
	ErrPasswordNeeded = errors.New("password required")
	ErrPasswordWrong = errors.New("wrong password")
)

type ShareContext struct {
	Link *domain.ShareLink
}

// Resolve loads the share by token and validates expiry/revocation.
// It does NOT enforce password — callers should check RequiresPassword separately.
// It does NOT consume one-time tokens — callers should call ConsumeIfOneTime after serving content.
func (s *ShareService) Resolve(ctx context.Context, token string) (*domain.ShareLink, error) {
	link, err := s.shares.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if link.RevokedAt != nil {
		return nil, ErrShareRevoked
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, ErrShareExpired
	}
	if link.OneTimeUse && link.ConsumedAt != nil {
		return nil, ErrShareConsumed
	}
	return link, nil
}

func (s *ShareService) VerifyPassword(link *domain.ShareLink, password string) error {
	if link.PasswordHash == "" {
		return nil
	}
	ok, err := auth.VerifyPassword(password, link.PasswordHash)
	if err != nil || !ok {
		return ErrPasswordWrong
	}
	return nil
}

// ConsumeIfOneTime marks a one-time share as consumed after a successful content serve.
// Safe to call on non-one-time shares (no-op).
func (s *ShareService) ConsumeIfOneTime(ctx context.Context, link *domain.ShareLink) error {
	if !link.OneTimeUse || link.ConsumedAt != nil {
		return nil
	}
	_, err := s.shares.ConsumeOneTime(ctx, link.ID)
	return err
}

// OpenFile returns a ReadCloser streaming the file content. The caller must close it.
// It also returns the file metadata for setting response headers.
func (s *ShareService) OpenFile(ctx context.Context, link *domain.ShareLink, rng *ports.Range) (io.ReadCloser, *domain.File, ports.ObjectInfo, error) {
	if link.TargetType != domain.ShareTargetFile {
		return nil, nil, ports.ObjectInfo{}, fmt.Errorf("%w: share is not a file", domain.ErrInvalidInput)
	}
	return s.storage.ProviderStreamByFileID(ctx, link.UserID, link.TargetID, rng)
}

// OpenFolderFile streams a file that's in the shared folder subtree.
// Verifies that the requested file is a descendant of the shared folder.
func (s *ShareService) OpenFolderFile(ctx context.Context, link *domain.ShareLink, fileID primitive.ObjectID, rng *ports.Range) (io.ReadCloser, *domain.File, ports.ObjectInfo, error) {
	if link.TargetType != domain.ShareTargetFolder {
		return nil, nil, ports.ObjectInfo{}, fmt.Errorf("%w: share is not a folder", domain.ErrInvalidInput)
	}
	f, err := s.files.Find(ctx, link.UserID, fileID)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	if f.FolderID == nil {
		return nil, nil, ports.ObjectInfo{}, domain.ErrForbidden
	}
	if !s.folderIsAncestor(ctx, link.UserID, link.TargetID, *f.FolderID) {
		return nil, nil, ports.ObjectInfo{}, domain.ErrForbidden
	}
	return s.storage.ProviderStreamByFileID(ctx, link.UserID, fileID, rng)
}

// ListFolder returns the subfolders and files inside a folder that's within the shared folder's subtree.
func (s *ShareService) ListFolder(ctx context.Context, link *domain.ShareLink, folderID primitive.ObjectID) ([]domain.Folder, []domain.File, error) {
	if link.TargetType != domain.ShareTargetFolder {
		return nil, nil, fmt.Errorf("%w: share is not a folder", domain.ErrInvalidInput)
	}
	if folderID != link.TargetID && !s.folderIsAncestor(ctx, link.UserID, link.TargetID, folderID) {
		return nil, nil, domain.ErrForbidden
	}
	children, err := s.folders.ListChildren(ctx, link.UserID, &folderID)
	if err != nil {
		return nil, nil, err
	}
	files, err := s.files.List(ctx, link.UserID, ports.FileListFilter{FolderID: &folderID, Limit: 500})
	if err != nil {
		return nil, nil, err
	}
	return children, files, nil
}

// folderIsAncestor checks whether ancestorID is the same as or an ancestor of candidateID.
func (s *ShareService) folderIsAncestor(ctx context.Context, userID, ancestorID, candidateID primitive.ObjectID) bool {
	if ancestorID == candidateID {
		return true
	}
	candidate, err := s.folders.Find(ctx, userID, candidateID)
	if err != nil {
		return false
	}
	return strings.Contains(candidate.Path, ","+ancestorID.Hex()+",")
}

func (s *ShareService) LogAccess(ctx context.Context, shareID primitive.ObjectID, ip, ua string, bytes int64) {
	_ = s.logs.Create(ctx, &domain.ShareAccessLog{
		ShareID:       shareID,
		IP:            ip,
		UserAgent:     ua,
		BytesStreamed: bytes,
	})
}

func generateShareToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
