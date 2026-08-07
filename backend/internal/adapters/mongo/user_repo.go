package mongoadapter

import (
	"context"
	"errors"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	col *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{col: db.Collection("users")}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now
	res, err := r.col.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var u domain.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) TouchLastLogin(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now().UTC()
	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"last_login_at": now, "updated_at": now}})
	return err
}

func (r *UserRepo) UpdateDisplayName(ctx context.Context, id primitive.ObjectID, displayName string) error {
	now := time.Now().UTC()
	res, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"display_name": displayName, "updated_at": now}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}
