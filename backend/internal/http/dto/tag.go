package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=50"`
	Color string `json:"color" validate:"omitempty,max=20"`
}

type UpdateTagRequest struct {
	Name  string `json:"name" validate:"omitempty,min=1,max=50"`
	Color string `json:"color" validate:"omitempty,max=20"`
}

type AttachTagRequest struct {
	TargetType string `json:"target_type" validate:"required,oneof=file folder"`
	TargetID   string `json:"target_id" validate:"required"`
}

type TagResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

func TagToResponse(t *domain.Tag) TagResponse {
	return TagResponse{
		ID:        t.ID.Hex(),
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: t.CreatedAt,
	}
}
