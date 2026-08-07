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

type APITokenRepo struct {
	col *mongo.Collection
}

func NewAPITokenRepo(db *mongo.Database) *APITokenRepo {
	return &APITokenRepo{col: db.Collection("api_tokens")}
}

func (r *APITokenRepo) Create(ctx context.Context, t *domain.APIToken) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.Scopes == nil {
		t.Scopes = []string{}
	}
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

func (r *APITokenRepo) FindByHash(ctx context.Context, hash string) (*domain.APIToken, error) {
	var t domain.APIToken
	err := r.col.FindOne(ctx, bson.M{"hash": hash}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *APITokenRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.APIToken, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.APIToken
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *APITokenRepo) Revoke(ctx context.Context, userID, id primitive.ObjectID) error {
	now := time.Now().UTC()
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "user_id": userID, "revoked_at": nil},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *APITokenRepo) TouchLastUsed(ctx context.Context, id primitive.ObjectID, when time.Time) error {
	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"last_used_at": when}})
	return err
}
