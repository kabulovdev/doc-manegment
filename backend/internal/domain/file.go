package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileStatus string

const (
	FileStatusUploading FileStatus = "uploading"
	FileStatusReady     FileStatus = "ready"
	FileStatusDeleted   FileStatus = "deleted"
)

type CustomFieldType string

const (
	CustomFieldText    CustomFieldType = "text"
	CustomFieldNumber  CustomFieldType = "number"
	CustomFieldDate    CustomFieldType = "date"
	CustomFieldBoolean CustomFieldType = "boolean"
)

type CustomField struct {
	Key   string          `bson:"key" json:"key"`
	Value string          `bson:"value" json:"value"`
	Type  CustomFieldType `bson:"type" json:"type"`
}

type MultipartPart struct {
	PartNumber int    `bson:"part_number"`
	ETag       string `bson:"etag"`
	Size       int64  `bson:"size"`
}

type MultipartState struct {
	UploadID string          `bson:"upload_id"`
	Parts    []MultipartPart `bson:"parts"`
}

// AIExtractionStatus tracks whether a file has been passed through a vision
// AI to produce a clean text representation for downstream tools.
type AIExtractionStatus string

const (
	AIExtractionNone    AIExtractionStatus = ""
	AIExtractionPending AIExtractionStatus = "pending"
	AIExtractionReady   AIExtractionStatus = "ready"
	AIExtractionFailed  AIExtractionStatus = "failed"
)

// AIExtraction is the cached result of a vision-capable chat provider reading
// a document. Written once per file (or after re-process), read by every
// downstream AI tool so they never touch raw bytes.
type AIExtraction struct {
	Status      AIExtractionStatus `bson:"status,omitempty" json:"status,omitempty"`
	Text        string             `bson:"text,omitempty" json:"text,omitempty"`
	Model       string             `bson:"model,omitempty" json:"model,omitempty"`
	Provider    string             `bson:"provider,omitempty" json:"provider,omitempty"`
	ProviderID  primitive.ObjectID `bson:"provider_id,omitempty" json:"provider_id,omitempty"`
	TokensIn    int                `bson:"tokens_in,omitempty" json:"tokens_in,omitempty"`
	TokensOut   int                `bson:"tokens_out,omitempty" json:"tokens_out,omitempty"`
	ExtractedAt *time.Time         `bson:"extracted_at,omitempty" json:"extracted_at,omitempty"`
	Error       string             `bson:"error,omitempty" json:"error,omitempty"`
}

type File struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty"`
	UserID         primitive.ObjectID  `bson:"user_id"`
	StorageID      primitive.ObjectID  `bson:"storage_id"`
	FolderID       *primitive.ObjectID `bson:"folder_id,omitempty"`
	Name           string              `bson:"name"`
	ObjectKey      string              `bson:"object_key"`
	SizeBytes      int64               `bson:"size_bytes"`
	MimeType       string              `bson:"mime_type"`
	ChecksumSHA256 string              `bson:"checksum_sha256,omitempty"`
	TagIDs         []primitive.ObjectID `bson:"tag_ids"`
	CustomFields   []CustomField       `bson:"custom_fields"`
	Multipart      *MultipartState     `bson:"multipart,omitempty"`
	Status         FileStatus          `bson:"status"`
	Starred        bool                `bson:"starred,omitempty"`
	AIExtraction   AIExtraction        `bson:"ai_extraction,omitempty"`
	UploadedAt     *time.Time          `bson:"uploaded_at,omitempty"`
	UpdatedAt      time.Time           `bson:"updated_at"`
	CreatedAt      time.Time           `bson:"created_at"`
}
