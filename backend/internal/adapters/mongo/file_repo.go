package mongoadapter

import (
	"context"
	"errors"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FileRepo struct {
	col *mongo.Collection
}

func NewFileRepo(db *mongo.Database) *FileRepo {
	return &FileRepo{col: db.Collection("files")}
}

func (r *FileRepo) Create(ctx context.Context, f *domain.File) error {
	now := time.Now().UTC()
	f.CreatedAt = now
	f.UpdatedAt = now
	if f.TagIDs == nil {
		f.TagIDs = []primitive.ObjectID{}
	}
	if f.CustomFields == nil {
		f.CustomFields = []domain.CustomField{}
	}
	res, err := r.col.InsertOne(ctx, f)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return err
	}
	f.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *FileRepo) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.File, error) {
	var f domain.File
	err := r.col.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepo) List(ctx context.Context, userID primitive.ObjectID, f ports.FileListFilter) ([]domain.File, error) {
	q := bson.M{"user_id": userID}
	if f.Status != nil {
		q["status"] = *f.Status
	} else {
		q["status"] = bson.M{"$ne": domain.FileStatusDeleted}
	}
	if f.FolderID != nil {
		q["folder_id"] = *f.FolderID
	}
	if f.StorageID != nil {
		q["storage_id"] = *f.StorageID
	}
	if f.TagID != nil {
		q["tag_ids"] = *f.TagID
	}
	if f.Starred != nil {
		if *f.Starred {
			q["starred"] = true
		} else {
			q["starred"] = bson.M{"$ne": true}
		}
	}
	if f.Query != "" {
		q["name"] = bson.M{"$regex": f.Query, "$options": "i"}
	}
	if f.FieldKey != "" {
		elem := bson.M{"key": f.FieldKey}
		if f.FieldValue != "" {
			elem["value"] = bson.M{"$regex": f.FieldValue, "$options": "i"}
		}
		q["custom_fields"] = bson.M{"$elemMatch": elem}
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	opts := options.Find().SetLimit(limit).SetSkip(f.Skip).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.col.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.File
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *FileRepo) Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error {
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

func (r *FileRepo) RawUpdate(ctx context.Context, userID, id primitive.ObjectID, ops bson.M) error {
	res, err := r.col.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, ops)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FileRepo) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	return r.Update(ctx, userID, id, bson.M{"status": domain.FileStatusDeleted})
}

func (r *FileRepo) HasFilesInStorage(ctx context.Context, storageID primitive.ObjectID) (bool, error) {
	n, err := r.col.CountDocuments(ctx, bson.M{
		"storage_id": storageID,
		"status":     bson.M{"$ne": domain.FileStatusDeleted},
	}, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *FileRepo) SoftDeleteByStorage(ctx context.Context, userID, storageID primitive.ObjectID) error {
	_, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID, "storage_id": storageID},
		bson.M{"$set": bson.M{"status": domain.FileStatusDeleted, "updated_at": time.Now().UTC()}},
	)
	return err
}

func (r *FileRepo) UpdateTagIDs(ctx context.Context, userID, id primitive.ObjectID, add, remove *primitive.ObjectID) error {
	ops := bson.M{"$set": bson.M{"updated_at": time.Now().UTC()}}
	if add != nil {
		ops["$addToSet"] = bson.M{"tag_ids": *add}
	}
	if remove != nil {
		ops["$pull"] = bson.M{"tag_ids": *remove}
	}
	res, err := r.col.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, ops)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FileRepo) RemoveTagFromAll(ctx context.Context, userID, tagID primitive.ObjectID) error {
	_, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID, "tag_ids": tagID},
		bson.M{"$pull": bson.M{"tag_ids": tagID}, "$set": bson.M{"updated_at": time.Now().UTC()}},
	)
	return err
}

func (r *FileRepo) SoftDeleteByFolderIDs(ctx context.Context, userID primitive.ObjectID, folderIDs []primitive.ObjectID) error {
	if len(folderIDs) == 0 {
		return nil
	}
	_, err := r.col.UpdateMany(ctx,
		bson.M{"user_id": userID, "folder_id": bson.M{"$in": folderIDs}},
		bson.M{"$set": bson.M{"status": domain.FileStatusDeleted, "updated_at": time.Now().UTC()}},
	)
	return err
}

func (r *FileRepo) SetStarred(ctx context.Context, userID, id primitive.ObjectID, starred bool) error {
	set := bson.M{"starred": starred, "updated_at": time.Now().UTC()}
	res, err := r.col.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FileRepo) HardDelete(ctx context.Context, userID, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FileRepo) ListDeletedOlderThan(ctx context.Context, before time.Time, limit int64) ([]domain.File, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	q := bson.M{
		"status":     domain.FileStatusDeleted,
		"updated_at": bson.M{"$lt": before},
	}
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "updated_at", Value: 1}})
	cur, err := r.col.Find(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.File
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
