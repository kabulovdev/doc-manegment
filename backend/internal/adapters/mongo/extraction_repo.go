package mongoadapter

import (
	"context"
	"errors"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExtractionRepo struct {
	col *mongo.Collection
}

func NewExtractionRepo(db *mongo.Database) *ExtractionRepo {
	return &ExtractionRepo{col: db.Collection("file_extractions")}
}

func (r *ExtractionRepo) FindByFile(ctx context.Context, userID, fileID primitive.ObjectID) (*domain.FileExtraction, error) {
	var e domain.FileExtraction
	err := r.col.FindOne(ctx, bson.M{"user_id": userID, "file_id": fileID}).Decode(&e)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExtractionRepo) Upsert(ctx context.Context, e *domain.FileExtraction) error {
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	_, err := r.col.UpdateOne(ctx,
		bson.M{"user_id": e.UserID, "file_id": e.FileID},
		bson.M{
			"$set": bson.M{
				"uploaded_at": e.UploadedAt,
				"mime":        e.Mime,
				"text":        e.Text,
				"char_count":  e.CharCount,
				"status":      e.Status,
				"engine":      e.Engine,
				"error":       e.Error,
				"duration_ms": e.DurationMS,
				"updated_at":  e.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"user_id":    e.UserID,
				"file_id":    e.FileID,
				"created_at": e.CreatedAt,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}
