package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/storage/s3compat"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"gitlab.com/docuemnt_manegment/backend/internal/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileService struct {
	files       ports.FileRepo
	storage     *StorageService
	maxFileSize int64
}

func NewFileService(files ports.FileRepo, storage *StorageService, maxFileSize int64) *FileService {
	return &FileService{files: files, storage: storage, maxFileSize: maxFileSize}
}

type InitUploadInput struct {
	StorageID    primitive.ObjectID
	FolderID     *primitive.ObjectID
	Name         string
	Size         int64
	Mime         string
	CustomFields []domain.CustomField
	TagIDs       []primitive.ObjectID
}

type InitUploadResult struct {
	FileID     primitive.ObjectID
	UploadID   string
	ObjectKey  string
	PartSize   int64
	Multipart  bool
}

const defaultPartSize int64 = 8 * 1024 * 1024

func (s *FileService) InitUpload(ctx context.Context, userID primitive.ObjectID, in InitUploadInput) (*InitUploadResult, error) {
	if in.Size <= 0 || in.Size > s.maxFileSize {
		return nil, fmt.Errorf("%w: invalid size", domain.ErrInvalidInput)
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	provider, err := s.storage.Provider(ctx, userID, in.StorageID)
	if err != nil {
		return nil, err
	}

	objectKey := util.ObjectKey(userID, name)
	if in.CustomFields == nil {
		in.CustomFields = []domain.CustomField{}
	}
	if in.TagIDs == nil {
		in.TagIDs = []primitive.ObjectID{}
	}

	f := &domain.File{
		UserID:       userID,
		StorageID:    in.StorageID,
		FolderID:     in.FolderID,
		Name:         name,
		ObjectKey:    objectKey,
		SizeBytes:    in.Size,
		MimeType:     in.Mime,
		CustomFields: in.CustomFields,
		TagIDs:       in.TagIDs,
		Status:       domain.FileStatusUploading,
	}

	res := &InitUploadResult{ObjectKey: objectKey, PartSize: defaultPartSize}
	if in.Size > defaultPartSize {
		uploadID, err := provider.InitiateMultipart(ctx, objectKey, in.Mime)
		if err != nil {
			return nil, err
		}
		f.Multipart = &domain.MultipartState{UploadID: uploadID, Parts: []domain.MultipartPart{}}
		res.UploadID = uploadID
		res.Multipart = true
	}

	if err := s.files.Create(ctx, f); err != nil {
		if f.Multipart != nil {
			_ = provider.AbortMultipart(ctx, objectKey, f.Multipart.UploadID)
		}
		return nil, err
	}
	res.FileID = f.ID
	return res, nil
}

func (s *FileService) UploadPart(ctx context.Context, userID, fileID primitive.ObjectID, partNumber int, body io.Reader, size int64) (string, error) {
	f, provider, err := s.loadUploading(ctx, userID, fileID)
	if err != nil {
		return "", err
	}
	if f.Multipart == nil {
		return "", fmt.Errorf("%w: file is not a multipart upload", domain.ErrInvalidInput)
	}
	if partNumber < 1 || partNumber > 10000 {
		return "", fmt.Errorf("%w: invalid part number", domain.ErrInvalidInput)
	}
	buf, err := io.ReadAll(io.LimitReader(body, size+1))
	if err != nil {
		return "", err
	}
	if int64(len(buf)) != size {
		return "", fmt.Errorf("%w: part size mismatch", domain.ErrInvalidInput)
	}
	etag, err := provider.UploadPart(ctx, f.ObjectKey, f.Multipart.UploadID, partNumber, bytes.NewReader(buf), size)
	if err != nil {
		return "", err
	}
	ops := bson.M{
		"$push": bson.M{
			"multipart.parts": domain.MultipartPart{PartNumber: partNumber, ETag: etag, Size: size},
		},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}
	if err := s.files.RawUpdate(ctx, userID, fileID, ops); err != nil {
		return "", err
	}
	return etag, nil
}

type CompleteInput struct {
	Parts []ports.Part
}

func (s *FileService) CompleteUpload(ctx context.Context, userID, fileID primitive.ObjectID, in CompleteInput) (*domain.File, error) {
	f, provider, err := s.loadUploading(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	if f.Multipart == nil {
		return nil, fmt.Errorf("%w: not multipart", domain.ErrInvalidInput)
	}
	if err := provider.CompleteMultipart(ctx, f.ObjectKey, f.Multipart.UploadID, in.Parts); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	set := bson.M{
		"status":       domain.FileStatusReady,
		"uploaded_at":  now,
		"multipart":    nil,
	}
	if err := s.files.Update(ctx, userID, fileID, set); err != nil {
		return nil, err
	}
	_ = s.storage.IncUsage(ctx, f.StorageID, f.SizeBytes, 1)
	return s.files.Find(ctx, userID, fileID)
}

func (s *FileService) AbortUpload(ctx context.Context, userID, fileID primitive.ObjectID) error {
	f, provider, err := s.loadUploading(ctx, userID, fileID)
	if err != nil {
		return err
	}
	if f.Multipart != nil {
		_ = provider.AbortMultipart(ctx, f.ObjectKey, f.Multipart.UploadID)
	} else {
		_ = provider.Delete(ctx, f.ObjectKey)
	}
	return s.files.Delete(ctx, userID, fileID)
}

// UploadSingle handles the non-multipart path. `size` must equal bytes in body.
func (s *FileService) UploadSingle(ctx context.Context, userID, fileID primitive.ObjectID, body io.Reader, size int64, mime string) (*domain.File, error) {
	f, provider, err := s.loadUploading(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	if f.Multipart != nil {
		return nil, fmt.Errorf("%w: file is a multipart upload", domain.ErrInvalidInput)
	}
	if mime == "" {
		mime = f.MimeType
	}
	buf, err := io.ReadAll(io.LimitReader(body, size+1))
	if err != nil {
		return nil, err
	}
	if int64(len(buf)) != size {
		return nil, fmt.Errorf("%w: body size mismatch", domain.ErrInvalidInput)
	}
	if len(buf) > 0 {
		head := buf
		if len(head) > 512 {
			head = head[:512]
		}
		sniffed := mimetype.Detect(head).String()
		if sniffed != "" && sniffed != "application/octet-stream" {
			mime = sniffed
		}
	}
	if err := provider.PutStream(ctx, f.ObjectKey, bytes.NewReader(buf), size, mime); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.files.Update(ctx, userID, fileID, bson.M{
		"status":      domain.FileStatusReady,
		"uploaded_at": now,
		"mime_type":   mime,
	}); err != nil {
		return nil, err
	}
	_ = s.storage.IncUsage(ctx, f.StorageID, size, 1)
	return s.files.Find(ctx, userID, fileID)
}

func (s *FileService) List(ctx context.Context, userID primitive.ObjectID, filter ports.FileListFilter) ([]domain.File, error) {
	return s.files.List(ctx, userID, filter)
}

func (s *FileService) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.File, error) {
	return s.files.Find(ctx, userID, id)
}

func (s *FileService) Stream(ctx context.Context, userID, id primitive.ObjectID, rng *ports.Range) (io.ReadCloser, *domain.File, ports.ObjectInfo, error) {
	f, err := s.files.Find(ctx, userID, id)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	if f.Status != domain.FileStatusReady {
		return nil, nil, ports.ObjectInfo{}, fmt.Errorf("%w: not ready", domain.ErrNotFound)
	}
	provider, err := s.storage.Provider(ctx, userID, f.StorageID)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	rc, info, err := provider.GetStream(ctx, f.ObjectKey, rng)
	if err != nil {
		return nil, nil, ports.ObjectInfo{}, err
	}
	return rc, f, info, nil
}

type UpdateInput struct {
	Name         *string
	TagIDs       *[]primitive.ObjectID
	CustomFields *[]domain.CustomField
}

func (s *FileService) Update(ctx context.Context, userID, id primitive.ObjectID, in UpdateInput) (*domain.File, error) {
	set := bson.M{}
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return nil, domain.ErrInvalidInput
		}
		set["name"] = n
	}
	if in.TagIDs != nil {
		set["tag_ids"] = *in.TagIDs
	}
	if in.CustomFields != nil {
		set["custom_fields"] = *in.CustomFields
	}
	if len(set) == 0 {
		return s.files.Find(ctx, userID, id)
	}
	if err := s.files.Update(ctx, userID, id, set); err != nil {
		return nil, err
	}
	return s.files.Find(ctx, userID, id)
}

func (s *FileService) SetStarred(ctx context.Context, userID, id primitive.ObjectID, starred bool) (*domain.File, error) {
	if err := s.files.SetStarred(ctx, userID, id, starred); err != nil {
		return nil, err
	}
	return s.files.Find(ctx, userID, id)
}

func (s *FileService) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	f, err := s.files.Find(ctx, userID, id)
	if err != nil {
		return err
	}
	if f.Status == domain.FileStatusDeleted {
		return nil
	}
	if err := s.files.Delete(ctx, userID, id); err != nil {
		return err
	}
	if f.Status == domain.FileStatusReady {
		_ = s.storage.IncUsage(ctx, f.StorageID, -f.SizeBytes, -1)
	}
	return nil
}

func (s *FileService) Restore(ctx context.Context, userID, id primitive.ObjectID) (*domain.File, error) {
	f, err := s.files.Find(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if f.Status != domain.FileStatusDeleted {
		return f, nil
	}
	if err := s.files.Update(ctx, userID, id, bson.M{"status": domain.FileStatusReady}); err != nil {
		return nil, err
	}
	_ = s.storage.IncUsage(ctx, f.StorageID, f.SizeBytes, 1)
	return s.files.Find(ctx, userID, id)
}

func (s *FileService) PermanentlyDelete(ctx context.Context, userID, id primitive.ObjectID) error {
	f, err := s.files.Find(ctx, userID, id)
	if err != nil {
		return err
	}
	provider, err := s.storage.Provider(ctx, userID, f.StorageID)
	if err == nil {
		_ = provider.Delete(ctx, f.ObjectKey)
	}
	if f.Status == domain.FileStatusReady {
		_ = s.storage.IncUsage(ctx, f.StorageID, -f.SizeBytes, -1)
	}
	return s.files.HardDelete(ctx, userID, id)
}

// PurgeOlderThan hard-deletes soft-deleted files whose updated_at is older than
// the given time. It iterates across all users, deleting the bucket object and
// the Mongo document. Returns the number of files purged and any errors
// collected; a per-file error does not abort the loop.
func (s *FileService) PurgeOlderThan(ctx context.Context, before time.Time, batch int64) (int, []error) {
	var errs []error
	purged := 0
	for {
		list, err := s.files.ListDeletedOlderThan(ctx, before, batch)
		if err != nil {
			errs = append(errs, fmt.Errorf("list deleted: %w", err))
			return purged, errs
		}
		if len(list) == 0 {
			return purged, errs
		}
		for i := range list {
			f := &list[i]
			if provider, err := s.storage.Provider(ctx, f.UserID, f.StorageID); err == nil {
				if err := provider.Delete(ctx, f.ObjectKey); err != nil {
					errs = append(errs, fmt.Errorf("delete object %s: %w", f.ObjectKey, err))
				}
			}
			if err := s.files.HardDelete(ctx, f.UserID, f.ID); err != nil {
				errs = append(errs, fmt.Errorf("hard delete %s: %w", f.ID.Hex(), err))
				continue
			}
			purged++
		}
		if int64(len(list)) < batch {
			return purged, errs
		}
	}
}

func (s *FileService) loadUploading(ctx context.Context, userID, fileID primitive.ObjectID) (*domain.File, *s3compat.Provider, error) {
	f, err := s.files.Find(ctx, userID, fileID)
	if err != nil {
		return nil, nil, err
	}
	if f.Status != domain.FileStatusUploading {
		return nil, nil, errors.New("file is not in uploading state")
	}
	provider, err := s.storage.Provider(ctx, userID, f.StorageID)
	if err != nil {
		return nil, nil, err
	}
	return f, provider, nil
}


