package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/adapters/storage/s3compat"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/storage/ssrf"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageService struct {
	repo       ports.StorageConfigRepo
	files      ports.FileRepo
	enc        ports.Encryptor
	ssrfPolicy ssrf.Policy

	clientCacheMu sync.Mutex
	clientCache   map[primitive.ObjectID]*s3compat.Provider
	httpClient    *http.Client
}

func NewStorageService(repo ports.StorageConfigRepo, files ports.FileRepo, enc ports.Encryptor, policy ssrf.Policy) *StorageService {
	transport := ssrf.GuardedTransport(policy)
	return &StorageService{
		repo:        repo,
		files:       files,
		enc:         enc,
		ssrfPolicy:  policy,
		clientCache: map[primitive.ObjectID]*s3compat.Provider{},
		httpClient:  &http.Client{Transport: transport, Timeout: policy.RequestTimeout},
	}
}

type CreateStorageInput struct {
	DisplayName    string
	Provider       domain.StorageProvider
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

func (s *StorageService) Create(ctx context.Context, userID primitive.ObjectID, in CreateStorageInput) (*domain.StorageConfig, error) {
	if !in.Provider.Valid() {
		return nil, domain.ErrInvalidInput
	}
	if strings.TrimSpace(in.DisplayName) == "" || strings.TrimSpace(in.Bucket) == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.AccessKey == "" || in.SecretKey == "" {
		return nil, domain.ErrInvalidInput
	}
	if _, err := ssrf.ValidateEndpoint(in.Endpoint, s.ssrfPolicy); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}

	accCT, accNonce, err := s.enc.Encrypt([]byte(in.AccessKey))
	if err != nil {
		return nil, err
	}
	secCT, secNonce, err := s.enc.Encrypt([]byte(in.SecretKey))
	if err != nil {
		return nil, err
	}

	sc := &domain.StorageConfig{
		UserID:         userID,
		DisplayName:    strings.TrimSpace(in.DisplayName),
		Provider:       in.Provider,
		Endpoint:       in.Endpoint,
		Region:         in.Region,
		Bucket:         strings.TrimSpace(in.Bucket),
		ForcePathStyle: in.ForcePathStyle,
		AccessKeyCT:    accCT,
		AccessKeyNonce: accNonce,
		SecretKeyCT:    secCT,
		SecretKeyNonce: secNonce,
		KeyVersion:     s.enc.KeyVersion(),
	}

	// Verify connectivity before persisting so users get immediate feedback.
	p, err := s3compat.NewFromStorageConfig(ctx, sc, in.AccessKey, in.SecretKey, s.httpClient)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	if err := p.HeadBucket(ctx); err != nil {
		return nil, fmt.Errorf("%w: bucket unreachable: %v", domain.ErrInvalidInput, err)
	}

	if err := s.repo.Create(ctx, sc); err != nil {
		return nil, err
	}
	return sc, nil
}

func (s *StorageService) List(ctx context.Context, userID primitive.ObjectID) ([]domain.StorageConfig, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *StorageService) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.StorageConfig, error) {
	return s.repo.Find(ctx, userID, id)
}

func (s *StorageService) Delete(ctx context.Context, userID, id primitive.ObjectID, force bool) error {
	has, err := s.files.HasFilesInStorage(ctx, id)
	if err != nil {
		return err
	}
	if has && !force {
		return fmt.Errorf("%w: storage has files; retry with force=true", domain.ErrConflict)
	}
	if has && force {
		if err := s.files.SoftDeleteByStorage(ctx, userID, id); err != nil {
			return err
		}
	}
	s.clientCacheMu.Lock()
	delete(s.clientCache, id)
	s.clientCacheMu.Unlock()
	return s.repo.Delete(ctx, userID, id)
}

func (s *StorageService) TestConnection(ctx context.Context, userID, id primitive.ObjectID) error {
	p, err := s.Provider(ctx, userID, id)
	if err != nil {
		return err
	}
	return p.HeadBucket(ctx)
}

func (s *StorageService) Resync(ctx context.Context, userID, id primitive.ObjectID) error {
	p, err := s.Provider(ctx, userID, id)
	if err != nil {
		return err
	}
	var totalBytes, totalCount int64
	cursor := ""
	for {
		page, err := p.List(ctx, "", cursor)
		if err != nil {
			_ = s.repo.Update(ctx, userID, id, bson.M{"last_error": err.Error(), "last_sync_at": time.Now().UTC()})
			return err
		}
		totalBytes += page.TotalBytes
		totalCount += page.TotalCount
		if !page.IsTruncated {
			break
		}
		cursor = page.NextCursor
	}
	return s.repo.SetUsage(ctx, id, totalBytes, totalCount, time.Now().UTC(), "")
}

func (s *StorageService) Provider(ctx context.Context, userID, id primitive.ObjectID) (*s3compat.Provider, error) {
	s.clientCacheMu.Lock()
	if p, ok := s.clientCache[id]; ok {
		s.clientCacheMu.Unlock()
		return p, nil
	}
	s.clientCacheMu.Unlock()

	sc, err := s.repo.Find(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	access, err := s.enc.Decrypt(sc.AccessKeyCT, sc.AccessKeyNonce)
	if err != nil {
		return nil, err
	}
	secret, err := s.enc.Decrypt(sc.SecretKeyCT, sc.SecretKeyNonce)
	if err != nil {
		return nil, err
	}
	p, err := s3compat.NewFromStorageConfig(ctx, sc, string(access), string(secret), s.httpClient)
	if err != nil {
		return nil, err
	}
	s.clientCacheMu.Lock()
	s.clientCache[id] = p
	s.clientCacheMu.Unlock()
	return p, nil
}

func (s *StorageService) IncUsage(ctx context.Context, id primitive.ObjectID, deltaBytes, deltaCount int64) error {
	return s.repo.IncrementUsage(ctx, id, deltaBytes, deltaCount)
}

// ProviderStreamByFileID is a convenience used by ShareService so it doesn't duplicate file lookup logic.
func (s *StorageService) ProviderStreamByFileID(ctx context.Context, userID, fileID primitive.ObjectID, rng *ports.Range) (io.ReadCloser, *domain.File, ports.ObjectInfo, error) {
	f, err := s.files.Find(ctx, userID, fileID)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	if f.Status != domain.FileStatusReady {
		return nil, nil, ports.ObjectInfo{}, domain.ErrNotFound
	}
	provider, err := s.Provider(ctx, userID, f.StorageID)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	rc, info, err := provider.GetStream(ctx, f.ObjectKey, rng)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	return rc, f, info, nil
}
