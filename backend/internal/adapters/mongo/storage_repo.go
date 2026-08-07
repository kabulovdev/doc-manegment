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

type StorageConfigRepo struct {
	col *mongo.Collection
}

func NewStorageConfigRepo(db *mongo.Database) *StorageConfigRepo {
	return &StorageConfigRepo{col: db.Collection("storage_configs")}
}

func (r *StorageConfigRepo) Create(ctx context.Context, s *domain.StorageConfig) error {
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	res, err := r.col.InsertOne(ctx, s)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return err
	}
	s.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *StorageConfigRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.StorageConfig, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.StorageConfig
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *StorageConfigRepo) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.StorageConfig, error) {
	var sc domain.StorageConfig
	err := r.col.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&sc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (r *StorageConfigRepo) Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error {
	set["updated_at"] = time.Now().UTC()
	res, err := r.col.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *StorageConfigRepo) IncrementUsage(ctx context.Context, id primitive.ObjectID, deltaBytes, deltaCount int64) error {
	_, err := r.col.UpdateByID(ctx, id, bson.M{
		"$inc": bson.M{"used_bytes": deltaBytes, "object_count": deltaCount},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	})
	return err
}

func (r *StorageConfigRepo) SetUsage(ctx context.Context, id primitive.ObjectID, totalBytes, totalCount int64, lastSync time.Time, lastErr string) error {
	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{
		"used_bytes":   totalBytes,
		"object_count": totalCount,
		"last_sync_at": lastSync,
		"last_error":   lastErr,
		"updated_at":   time.Now().UTC(),
	}})
	return err
}

func (r *StorageConfigRepo) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}
