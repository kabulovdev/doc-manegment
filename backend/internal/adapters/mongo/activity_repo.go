package mongoadapter

import (
	"context"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ActivityLogRepo struct {
	col *mongo.Collection
}

func NewActivityLogRepo(db *mongo.Database) *ActivityLogRepo {
	return &ActivityLogRepo{col: db.Collection("activity_logs")}
}

func (r *ActivityLogRepo) Create(ctx context.Context, l *domain.ActivityLog) error {
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	res, err := r.col.InsertOne(ctx, l)
	if err != nil {
		return err
	}
	l.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ActivityLogRepo) List(ctx context.Context, f ports.ActivityFilter) ([]domain.ActivityLog, error) {
	q := bson.M{"user_id": f.UserID}
	if f.SubjectID != nil {
		q["subject_id"] = *f.SubjectID
	}
	if f.SubjectType != nil {
		q["subject_type"] = *f.SubjectType
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.col.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.ActivityLog
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
