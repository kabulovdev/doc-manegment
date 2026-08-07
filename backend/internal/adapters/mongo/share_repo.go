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

type ShareRepo struct {
	col *mongo.Collection
}

func NewShareRepo(db *mongo.Database) *ShareRepo {
	return &ShareRepo{col: db.Collection("share_links")}
}

func (r *ShareRepo) Create(ctx context.Context, s *domain.ShareLink) error {
	s.CreatedAt = time.Now().UTC()
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

func (r *ShareRepo) FindByToken(ctx context.Context, token string) (*domain.ShareLink, error) {
	var s domain.ShareLink
	err := r.col.FindOne(ctx, bson.M{"token": token}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShareRepo) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.ShareLink, error) {
	var s domain.ShareLink
	err := r.col.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShareRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.ShareLink, error) {
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.ShareLink
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ShareRepo) Revoke(ctx context.Context, userID, id primitive.ObjectID) error {
	now := time.Now().UTC()
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "user_id": userID},
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

// ConsumeOneTime atomically flips consumed_at from null to now when the token is one-time.
// Returns true if we were the first to consume it.
func (r *ShareRepo) ConsumeOneTime(ctx context.Context, id primitive.ObjectID) (bool, error) {
	now := time.Now().UTC()
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "one_time_use": true, "consumed_at": nil},
		bson.M{"$set": bson.M{"consumed_at": now}},
	)
	if err != nil {
		return false, err
	}
	return res.ModifiedCount == 1, nil
}

type ShareAccessLogRepo struct {
	col *mongo.Collection
}

func NewShareAccessLogRepo(db *mongo.Database) *ShareAccessLogRepo {
	return &ShareAccessLogRepo{col: db.Collection("share_access_logs")}
}

func (r *ShareAccessLogRepo) Create(ctx context.Context, l *domain.ShareAccessLog) error {
	if l.AccessedAt.IsZero() {
		l.AccessedAt = time.Now().UTC()
	}
	_, err := r.col.InsertOne(ctx, l)
	return err
}

func (r *ShareAccessLogRepo) TailByShare(ctx context.Context, shareID primitive.ObjectID, limit int64) ([]domain.ShareAccessLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	cur, err := r.col.Find(ctx,
		bson.M{"share_id": shareID},
		options.Find().SetSort(bson.D{{Key: "accessed_at", Value: -1}}).SetLimit(limit),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.ShareAccessLog
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
