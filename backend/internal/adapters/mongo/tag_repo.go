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

type TagRepo struct {
	col *mongo.Collection
}

func NewTagRepo(db *mongo.Database) *TagRepo {
	return &TagRepo{col: db.Collection("tags")}
}

func (r *TagRepo) Create(ctx context.Context, t *domain.Tag) error {
	t.CreatedAt = time.Now().UTC()
	res, err := r.col.InsertOne(ctx, t)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return err
	}
	t.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TagRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.Tag, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Tag
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TagRepo) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.Tag, error) {
	var t domain.Tag
	err := r.col.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepo) Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error {
	res, err := r.col.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TagRepo) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}
