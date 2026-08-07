package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CustomFieldDTO struct {
	Key   string `json:"key" validate:"required,max=100"`
	Value string `json:"value" validate:"max=1000"`
	Type  string `json:"type" validate:"required,oneof=text number date boolean"`
}

type InitUploadRequest struct {
	StorageID    string           `json:"storage_id" validate:"required"`
	FolderID     *string          `json:"folder_id,omitempty"`
	Name         string           `json:"name" validate:"required,min=1,max=255"`
	Size         int64            `json:"size" validate:"required,min=1"`
	Mime         string           `json:"mime"`
	CustomFields []CustomFieldDTO `json:"custom_fields" validate:"dive"`
	TagIDs       []string         `json:"tag_ids"`
}

type InitUploadResponse struct {
	FileID    string `json:"file_id"`
	UploadID  string `json:"upload_id,omitempty"`
	ObjectKey string `json:"object_key"`
	PartSize  int64  `json:"part_size"`
	Multipart bool   `json:"multipart"`
}

type UpdateFileRequest struct {
	Name         *string           `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	TagIDs       *[]string         `json:"tag_ids,omitempty"`
	CustomFields *[]CustomFieldDTO `json:"custom_fields,omitempty" validate:"omitempty,dive"`
}

type CompletePartDTO struct {
	PartNumber int    `json:"part_number" validate:"required,min=1"`
	ETag       string `json:"etag" validate:"required"`
}

type CompleteRequest struct {
	Parts []CompletePartDTO `json:"parts" validate:"required,min=1,dive"`
}

type FileResponse struct {
	ID              string           `json:"id"`
	StorageID       string           `json:"storage_id"`
	FolderID        *string          `json:"folder_id,omitempty"`
	Name            string           `json:"name"`
	SizeBytes       int64            `json:"size_bytes"`
	MimeType        string           `json:"mime_type"`
	Status          string           `json:"status"`
	Starred         bool             `json:"starred"`
	TagIDs          []string         `json:"tag_ids"`
	CustomFields    []CustomFieldDTO `json:"custom_fields"`
	AIStatus        string           `json:"ai_status,omitempty"`
	UploadedAt      *time.Time       `json:"uploaded_at,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}

type StarRequest struct {
	Starred bool `json:"starred"`
}

func FileToResponse(f *domain.File) FileResponse {
	tagIDs := make([]string, 0, len(f.TagIDs))
	for _, id := range f.TagIDs {
		tagIDs = append(tagIDs, id.Hex())
	}
	fields := make([]CustomFieldDTO, 0, len(f.CustomFields))
	for _, cf := range f.CustomFields {
		fields = append(fields, CustomFieldDTO{Key: cf.Key, Value: cf.Value, Type: string(cf.Type)})
	}
	var folderID *string
	if f.FolderID != nil {
		h := f.FolderID.Hex()
		folderID = &h
	}
	aiStatus := string(f.AIExtraction.Status)
	if aiStatus == "" {
		aiStatus = "none"
	}
	return FileResponse{
		ID:           f.ID.Hex(),
		StorageID:    f.StorageID.Hex(),
		FolderID:     folderID,
		Name:         f.Name,
		SizeBytes:    f.SizeBytes,
		MimeType:     f.MimeType,
		Status:       string(f.Status),
		Starred:      f.Starred,
		TagIDs:       tagIDs,
		CustomFields: fields,
		AIStatus:     aiStatus,
		UploadedAt:   f.UploadedAt,
		CreatedAt:    f.CreatedAt,
	}
}
