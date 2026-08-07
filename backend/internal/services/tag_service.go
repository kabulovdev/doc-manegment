package services

import (
	"context"
	"strings"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TagService struct {
	tags    ports.TagRepo
	files   ports.FileRepo
	folders ports.FolderRepo
}

func NewTagService(tags ports.TagRepo, files ports.FileRepo, folders ports.FolderRepo) *TagService {
	return &TagService{tags: tags, files: files, folders: folders}
}

func (s *TagService) Create(ctx context.Context, userID primitive.ObjectID, name, color string) (*domain.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}
	if color == "" {
		color = "#64748b" // slate-500
	}
	t := &domain.Tag{UserID: userID, Name: name, Color: color}
	if err := s.tags.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TagService) List(ctx context.Context, userID primitive.ObjectID) ([]domain.Tag, error) {
	return s.tags.ListByUser(ctx, userID)
}

func (s *TagService) Update(ctx context.Context, userID, id primitive.ObjectID, name, color string) error {
	set := bson.M{}
	if name != "" {
		set["name"] = strings.TrimSpace(name)
	}
	if color != "" {
		set["color"] = color
	}
	if len(set) == 0 {
		return domain.ErrInvalidInput
	}
	return s.tags.Update(ctx, userID, id, set)
}

func (s *TagService) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	if _, err := s.tags.Find(ctx, userID, id); err != nil {
		return err
	}
	if err := s.files.RemoveTagFromAll(ctx, userID, id); err != nil {
		return err
	}
	return s.tags.Delete(ctx, userID, id)
}

func (s *TagService) Attach(ctx context.Context, userID, tagID primitive.ObjectID, target domain.TagTargetType, targetID primitive.ObjectID) error {
	if _, err := s.tags.Find(ctx, userID, tagID); err != nil {
		return err
	}
	switch target {
	case domain.TagTargetFile:
		return s.files.UpdateTagIDs(ctx, userID, targetID, &tagID, nil)
	case domain.TagTargetFolder:
		return s.folders.UpdateTagIDs(ctx, userID, targetID, &tagID, nil)
	default:
		return domain.ErrInvalidInput
	}
}

func (s *TagService) Detach(ctx context.Context, userID, tagID primitive.ObjectID, target domain.TagTargetType, targetID primitive.ObjectID) error {
	switch target {
	case domain.TagTargetFile:
		return s.files.UpdateTagIDs(ctx, userID, targetID, nil, &tagID)
	case domain.TagTargetFolder:
		return s.folders.UpdateTagIDs(ctx, userID, targetID, nil, &tagID)
	default:
		return domain.ErrInvalidInput
	}
}
