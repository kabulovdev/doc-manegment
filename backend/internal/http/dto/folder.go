package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CreateFolderRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	ParentID *string `json:"parent_id,omitempty"`
}

type RenameFolderRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type MoveFolderRequest struct {
	NewParentID *string `json:"new_parent_id,omitempty"`
}

type FolderResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *string   `json:"parent_id,omitempty"`
	Depth     int       `json:"depth"`
	TagIDs    []string  `json:"tag_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FolderListResponse struct {
	Folder      *FolderResponse  `json:"folder,omitempty"`
	Breadcrumbs []FolderResponse `json:"breadcrumbs,omitempty"`
	Children    []FolderResponse `json:"children"`
}

func FolderToResponse(f *domain.Folder) FolderResponse {
	tagIDs := make([]string, 0, len(f.TagIDs))
	for _, id := range f.TagIDs {
		tagIDs = append(tagIDs, id.Hex())
	}
	var parentID *string
	if f.ParentID != nil {
		s := f.ParentID.Hex()
		parentID = &s
	}
	return FolderResponse{
		ID:        f.ID.Hex(),
		Name:      f.Name,
		ParentID:  parentID,
		Depth:     f.Depth,
		TagIDs:    tagIDs,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}
