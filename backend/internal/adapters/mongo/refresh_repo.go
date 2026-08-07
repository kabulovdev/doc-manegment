package mongoadapter

import (
	"context"
	"errors"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RefreshTokenRepo struct {
	col *mongo.Collection
}

func NewRefreshTokenRepo(db *mongo.Database) *RefreshTokenRepo {
	return &RefreshTokenRepo{col: db.Collection("refresh_tokens")}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	res, err := r.col.InsertOne(ctx, t)
	if err != nil {
		return err
	}
	t.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *RefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := r.col.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *RefreshTokenRepo) MarkReplaced(ctx context.Context, id, replacedBy primitive.ObjectID) error {
	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"replaced_by": replacedBy}})
	return err
}

func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, familyID primitive.ObjectID) error {
	_, err := r.col.UpdateMany(ctx, bson.M{"family_id": familyID}, bson.M{"$set": bson.M{"revoked": true}})
	return err
}
