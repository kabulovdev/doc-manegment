package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShareTargetType string

const (
	ShareTargetFile   ShareTargetType = "file"
	ShareTargetFolder ShareTargetType = "folder"
)

type ShareLink struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	Token        string             `bson:"token"`
	TargetType   ShareTargetType    `bson:"target_type"`
	TargetID     primitive.ObjectID `bson:"target_id"`
	ExpiresAt    *time.Time         `bson:"expires_at,omitempty"`
	PasswordHash string             `bson:"password_hash,omitempty"`
	OneTimeUse   bool               `bson:"one_time_use"`
	ConsumedAt   *time.Time         `bson:"consumed_at,omitempty"`
	RevokedAt    *time.Time         `bson:"revoked_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at"`
}

type ShareAccessLog struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	ShareID       primitive.ObjectID `bson:"share_id"`
	IP            string             `bson:"ip,omitempty"`
	UserAgent     string             `bson:"user_agent,omitempty"`
	AccessedAt    time.Time          `bson:"accessed_at"`
	BytesStreamed int64              `bson:"bytes_streamed,omitempty"`
}
