package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExtractionStatus tracks the lifecycle of an extraction attempt. A single
// FileExtraction is kept per (file_id, uploaded_at); re-uploads invalidate
// the cache by matching against UploadedAt.
type ExtractionStatus string

const (
	ExtractionPending     ExtractionStatus = "pending"
	ExtractionReady       ExtractionStatus = "ready"
	ExtractionUnsupported ExtractionStatus = "unsupported"
	ExtractionError       ExtractionStatus = "error"
)

type FileExtraction struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id"`
	FileID     primitive.ObjectID `bson:"file_id"`
	UploadedAt *time.Time         `bson:"uploaded_at,omitempty"`
	Mime       string             `bson:"mime,omitempty"`

	Text        string           `bson:"text,omitempty"`
	CharCount   int              `bson:"char_count"`
	Status      ExtractionStatus `bson:"status"`
	Engine      string           `bson:"engine,omitempty"`
	Error       string           `bson:"error,omitempty"`
	DurationMS  int64            `bson:"duration_ms,omitempty"`
	CreatedAt   time.Time        `bson:"created_at"`
	UpdatedAt   time.Time        `bson:"updated_at"`
}
