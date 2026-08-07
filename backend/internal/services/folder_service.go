package services

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FolderService struct {
	folders ports.FolderRepo
	files   ports.FileRepo
}

func NewFolderService(folders ports.FolderRepo, files ports.FileRepo) *FolderService {
	return &FolderService{folders: folders, files: files}
}

func (s *FolderService) Create(ctx context.Context, userID primitive.ObjectID, name string, parentID *primitive.ObjectID) (*domain.Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}
	path := ","
	depth := 0
	if parentID != nil {
		parent, err := s.folders.Find(ctx, userID, *parentID)
		if err != nil {
			return nil, err
		}
		path = parent.Path + parent.ID.Hex() + ","
		depth = parent.Depth + 1
	}
	f := &domain.Folder{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
		Path:     path,
		Depth:    depth,
	}
	if err := s.folders.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FolderService) Rename(ctx context.Context, userID, id primitive.ObjectID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.ErrInvalidInput
	}
	return s.folders.Rename(ctx, userID, id, name)
}

func (s *FolderService) Move(ctx context.Context, userID, id primitive.ObjectID, newParentID *primitive.ObjectID) error {
	f, err := s.folders.Find(ctx, userID, id)
	if err != nil {
		return err
	}
	oldPrefix := f.Path + f.ID.Hex() + ","
	var newParentPath string
	newDepth := 0
	if newParentID != nil {
		if *newParentID == id {
			return fmt.Errorf("%w: cannot move folder into itself", domain.ErrInvalidInput)
		}
		newParent, err := s.folders.Find(ctx, userID, *newParentID)
		if err != nil {
			return err
		}
		// Prevent moving a folder into one of its own descendants.
		if strings.HasPrefix(newParent.Path, oldPrefix) {
			return fmt.Errorf("%w: cannot move folder into its descendant", domain.ErrInvalidInput)
		}
		newParentPath = newParent.Path + newParent.ID.Hex() + ","
		newDepth = newParent.Depth + 1
	} else {
		newParentPath = ","
	}
	// Update the folder itself first, then rewrite descendants.
	if err := s.folders.UpdatePathAndParent(ctx, userID, id, newParentID, newParentPath, newDepth); err != nil {
		return err
	}
	newPrefix := newParentPath + id.Hex() + ","
	depthDelta := newDepth - f.Depth
	return s.folders.RewriteSubtreePath(ctx, userID, oldPrefix, newPrefix, depthDelta)
}

func (s *FolderService) Delete(ctx context.Context, userID, id primitive.ObjectID, recursive bool) error {
	f, err := s.folders.Find(ctx, userID, id)
	if err != nil {
		return err
	}
	// Check for direct children and files.
	children, err := s.folders.ListChildren(ctx, userID, &f.ID)
	if err != nil {
		return err
	}
	files, err := s.files.List(ctx, userID, ports.FileListFilter{FolderID: &f.ID, Limit: 1})
	if err != nil {
		return err
	}
	hasContent := len(children) > 0 || len(files) > 0
	if hasContent && !recursive {
		return fmt.Errorf("%w: folder is not empty; retry with recursive=true", domain.ErrConflict)
	}
	if !hasContent {
		return s.folders.DeleteSubtree(ctx, userID, id)
	}
	// Recursive: collect all descendant folder IDs (plus self), soft-delete all files in them, then delete folders.
	ids, err := s.folders.ListDescendantIDs(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := s.files.SoftDeleteByFolderIDs(ctx, userID, ids); err != nil {
		return err
	}
	return s.folders.DeleteSubtree(ctx, userID, id)
}

func (s *FolderService) List(ctx context.Context, userID primitive.ObjectID, parentID *primitive.ObjectID) ([]domain.Folder, error) {
	return s.folders.ListChildren(ctx, userID, parentID)
}

func (s *FolderService) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.Folder, error) {
	return s.folders.Find(ctx, userID, id)
}

// Breadcrumbs returns the folder's ancestors (root → ... → folder) in order.
func (s *FolderService) Breadcrumbs(ctx context.Context, userID, id primitive.ObjectID) ([]domain.Folder, error) {
	f, err := s.folders.Find(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	ids := parsePathIDs(f.Path)
	if len(ids) == 0 {
		return []domain.Folder{*f}, nil
	}
	ancestors, err := s.folders.ListByIDs(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[primitive.ObjectID]domain.Folder, len(ancestors))
	for _, a := range ancestors {
		byID[a.ID] = a
	}
	ordered := make([]domain.Folder, 0, len(ids)+1)
	for _, id := range ids {
		if a, ok := byID[id]; ok {
			ordered = append(ordered, a)
		}
	}
	ordered = append(ordered, *f)
	return ordered, nil
}

func parsePathIDs(path string) []primitive.ObjectID {
	parts := strings.Split(strings.Trim(path, ","), ",")
	out := make([]primitive.ObjectID, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		id, err := primitive.ObjectIDFromHex(p)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out
}
