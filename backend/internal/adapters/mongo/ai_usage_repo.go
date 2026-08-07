package mongoadapter

import (
	"context"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AIUsageRepo struct {
	col *mongo.Collection
}

func NewAIUsageRepo(db *mongo.Database) *AIUsageRepo {
	return &AIUsageRepo{col: db.Collection("ai_usage")}
}

func (r *AIUsageRepo) Create(ctx context.Context, u *domain.AIUsage) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	res, err := r.col.InsertOne(ctx, u)
	if err != nil {
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// TotalsByConfig aggregates tokens_in, tokens_out, and call count per AIConfig
// for a single user. Used by the settings UI to display per-provider usage.
func (r *AIUsageRepo) TotalsByConfig(ctx context.Context, userID primitive.ObjectID) (map[primitive.ObjectID]ports.AIUsageTotals, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"user_id": userID}},
		bson.M{"$group": bson.M{
			"_id":        "$ai_config_id",
			"tokens_in":  bson.M{"$sum": "$tokens_in"},
			"tokens_out": bson.M{"$sum": "$tokens_out"},
			"calls":      bson.M{"$sum": 1},
		}},
	}
	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := map[primitive.ObjectID]ports.AIUsageTotals{}
	for cur.Next(ctx) {
		var row struct {
			ID        primitive.ObjectID `bson:"_id"`
			TokensIn  int64              `bson:"tokens_in"`
			TokensOut int64              `bson:"tokens_out"`
			Calls     int64              `bson:"calls"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out[row.ID] = ports.AIUsageTotals{
			TokensIn:  row.TokensIn,
			TokensOut: row.TokensOut,
			Calls:     row.Calls,
		}
	}
	return out, cur.Err()
}
