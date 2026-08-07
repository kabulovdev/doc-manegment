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

type AIConfigRepo struct {
	col *mongo.Collection
}

func NewAIConfigRepo(db *mongo.Database) *AIConfigRepo {
	return &AIConfigRepo{col: db.Collection("ai_configs")}
}

func (r *AIConfigRepo) Create(ctx context.Context, c *domain.AIConfig) error {
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.Capabilities == nil {
		c.Capabilities = []domain.AICapability{}
	}
	res, err := r.col.InsertOne(ctx, c)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return err
	}
	c.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AIConfigRepo) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.AIConfig, error) {
	var c domain.AIConfig
	err := r.col.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *AIConfigRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.AIConfig, error) {
	cur, err := r.col.Find(ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.AIConfig
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AIConfigRepo) Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error {
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

// ClearDefault unsets the per-capability default flag on every config belonging
// to the user. Call this before setting a new default so the invariant "at
// most one default per capability" holds.
func (r *AIConfigRepo) ClearDefault(ctx context.Context, userID primitive.ObjectID, capability domain.AICapability) error {
	field, ok := defaultFieldFor(capability)
	if !ok {
		return domain.ErrInvalidInput
	}
	_, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID, field: true},
		bson.M{"$set": bson.M{field: false, "updated_at": time.Now().UTC()}},
	)
	return err
}

func (r *AIConfigRepo) IncrementUsage(ctx context.Context, id primitive.ObjectID, tokensIn, tokensOut int64, when time.Time) error {
	_, err := r.col.UpdateByID(ctx, id, bson.M{
		"$inc": bson.M{"used_tokens_in": tokensIn, "used_tokens_out": tokensOut},
		"$set": bson.M{"last_used_at": when, "updated_at": time.Now().UTC()},
	})
	return err
}

func (r *AIConfigRepo) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func defaultFieldFor(c domain.AICapability) (string, bool) {
	switch c {
	case domain.AICapChat:
		return "is_default_chat", true
	case domain.AICapEmbed:
		return "is_default_embed", true
	case domain.AICapTranscribe:
		return "is_default_transcribe", true
	}
	return "", false
}
