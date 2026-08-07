package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CreateShareRequest struct {
	TargetType string     `json:"target_type" validate:"required,oneof=file folder"`
	TargetID   string     `json:"target_id" validate:"required"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Password   string     `json:"password,omitempty" validate:"omitempty,min=4,max=128"`
	OneTimeUse bool       `json:"one_time_use"`
}

type UnlockShareRequest struct {
	Password string `json:"password" validate:"required"`
}

type ShareResponse struct {
	ID              string     `json:"id"`
	Token           string     `json:"token"`
	TargetType      string     `json:"target_type"`
	TargetID        string     `json:"target_id"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	HasPassword     bool       `json:"has_password"`
	OneTimeUse      bool       `json:"one_time_use"`
	Consumed        bool       `json:"consumed"`
	Revoked         bool       `json:"revoked"`
	CreatedAt       time.Time  `json:"created_at"`
}

type PublicShareMetadata struct {
	TargetType      string    `json:"target_type"`
	RequiresPassword bool     `json:"requires_password"`
	Name             string   `json:"name"`
	MimeType         string   `json:"mime_type,omitempty"`
	SizeBytes        int64    `json:"size_bytes,omitempty"`
	FolderID         string   `json:"folder_id,omitempty"`
	Unlocked         bool     `json:"unlocked"`
}

type PublicFolderListing struct {
	Folder   *PublicFolderEntry   `json:"folder"`
	Children []PublicFolderEntry  `json:"children"`
	Files    []PublicFileEntry    `json:"files"`
}

type PublicFolderEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublicFileEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type ShareAccessLogEntry struct {
	AccessedAt    time.Time `json:"accessed_at"`
	IP            string    `json:"ip,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
	BytesStreamed int64     `json:"bytes_streamed,omitempty"`
}

func ShareToResponse(s *domain.ShareLink) ShareResponse {
	return ShareResponse{
		ID:          s.ID.Hex(),
		Token:       s.Token,
		TargetType:  string(s.TargetType),
		TargetID:    s.TargetID.Hex(),
		ExpiresAt:   s.ExpiresAt,
		HasPassword: s.PasswordHash != "",
		OneTimeUse:  s.OneTimeUse,
		Consumed:    s.ConsumedAt != nil,
		Revoked:     s.RevokedAt != nil,
		CreatedAt:   s.CreatedAt,
	}
}
