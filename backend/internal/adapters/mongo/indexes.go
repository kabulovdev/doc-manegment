package mongoadapter

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if _, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	if _, err := db.Collection("refresh_tokens").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "token_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("storage_configs").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "display_name", Value: 1}}, Options: options.Index().SetUnique(true)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("folders").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "parent_id", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "path", Value: 1}}},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("files").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "folder_id", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "tag_ids", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "custom_fields.key", Value: 1}, {Key: "custom_fields.value", Value: 1}}},
		{Keys: bson.D{{Key: "storage_id", Value: 1}, {Key: "object_key", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "starred", Value: 1}, {Key: "updated_at", Value: -1}}, Options: options.Index().SetSparse(true)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("tags").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	if _, err := db.Collection("share_links").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "token", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("share_access_logs").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "share_id", Value: 1}, {Key: "accessed_at", Value: -1}}},
		{Keys: bson.D{{Key: "accessed_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(60 * 60 * 24 * 90)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("activity_logs").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "subject_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(60 * 60 * 24 * 180)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("api_tokens").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("ai_configs").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "display_name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_default_chat", Value: 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_default_embed", Value: 1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_default_transcribe", Value: 1}}, Options: options.Index().SetSparse(true)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("ai_usage").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "ai_config_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(60 * 60 * 24 * 180)},
	}); err != nil {
		return err
	}

	if _, err := db.Collection("file_extractions").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "file_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	return nil
}
