package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CreateStorageRequest struct {
	DisplayName    string `json:"display_name" validate:"required,min=1,max=100"`
	Provider       string `json:"provider" validate:"required,oneof=s3 r2 minio"`
	Endpoint       string `json:"endpoint" validate:"required,url"`
	Region         string `json:"region"`
	Bucket         string `json:"bucket" validate:"required,min=1,max=100"`
	AccessKey      string `json:"access_key" validate:"required"`
	SecretKey      string `json:"secret_key" validate:"required"`
	ForcePathStyle bool   `json:"force_path_style"`
}

type StorageResponse struct {
	ID             string     `json:"id"`
	DisplayName    string     `json:"display_name"`
	Provider       string     `json:"provider"`
	Endpoint       string     `json:"endpoint"`
	Region         string     `json:"region"`
	Bucket         string     `json:"bucket"`
	ForcePathStyle bool       `json:"force_path_style"`
	UsedBytes      int64      `json:"used_bytes"`
	ObjectCount    int64      `json:"object_count"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func StorageToResponse(s *domain.StorageConfig) StorageResponse {
	return StorageResponse{
		ID:             s.ID.Hex(),
		DisplayName:    s.DisplayName,
		Provider:       string(s.Provider),
		Endpoint:       s.Endpoint,
		Region:         s.Region,
		Bucket:         s.Bucket,
		ForcePathStyle: s.ForcePathStyle,
		UsedBytes:      s.UsedBytes,
		ObjectCount:    s.ObjectCount,
		LastSyncAt:     s.LastSyncAt,
		LastError:      s.LastError,
		CreatedAt:      s.CreatedAt,
	}
}
