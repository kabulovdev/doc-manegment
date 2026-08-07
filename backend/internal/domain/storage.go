package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageProvider string

const (
	ProviderS3    StorageProvider = "s3"
	ProviderR2    StorageProvider = "r2"
	ProviderMinio StorageProvider = "minio"
)

func (p StorageProvider) Valid() bool {
	switch p {
	case ProviderS3, ProviderR2, ProviderMinio:
		return true
	}
	return false
}

type StorageConfig struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         primitive.ObjectID `bson:"user_id"`
	DisplayName    string             `bson:"display_name"`
	Provider       StorageProvider    `bson:"provider"`
	Endpoint       string             `bson:"endpoint"`
	Region         string             `bson:"region"`
	Bucket         string             `bson:"bucket"`
	ForcePathStyle bool               `bson:"force_path_style"`
	AccessKeyCT    []byte             `bson:"access_key_ct"`
	AccessKeyNonce []byte             `bson:"access_key_nonce"`
	SecretKeyCT    []byte             `bson:"secret_key_ct"`
	SecretKeyNonce []byte             `bson:"secret_key_nonce"`
	KeyVersion     int                `bson:"key_version"`
	UsedBytes      int64              `bson:"used_bytes"`
	ObjectCount    int64              `bson:"object_count"`
	LastSyncAt     *time.Time         `bson:"last_sync_at,omitempty"`
	LastError      string             `bson:"last_error,omitempty"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}
