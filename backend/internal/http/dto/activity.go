package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type ActivityResponse struct {
	ID          string         `json:"id"`
	SubjectType string         `json:"subject_type"`
	SubjectID   string         `json:"subject_id"`
	Action      string         `json:"action"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

func ActivityToResponse(l *domain.ActivityLog) ActivityResponse {
	return ActivityResponse{
		ID:          l.ID.Hex(),
		SubjectType: string(l.SubjectType),
		SubjectID:   l.SubjectID.Hex(),
		Action:      string(l.Action),
		Metadata:    l.Metadata,
		CreatedAt:   l.CreatedAt,
	}
}
