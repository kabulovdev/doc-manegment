package mongoadapter

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FolderRepo struct {
	col *mongo.Collection
}

func NewFolderRepo(db *mongo.Database) *FolderRepo {
	return &FolderRepo{col: db.Collection("folders")}
}

func (r *FolderRepo) Create(ctx context.Context, f *domain.Folder) error {
	now := time.Now().UTC()
	f.CreatedAt = now
	f.UpdatedAt = now
	if f.TagIDs == nil {
		f.TagIDs = []primitive.ObjectID{}
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

func (r *FolderRepo) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.Folder, error) {
	var f domain.Folder
	err := r.col.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&f)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FolderRepo) ListChildren(ctx context.Context, userID primitive.ObjectID, parentID *primitive.ObjectID) ([]domain.Folder, error) {
	q := bson.M{"user_id": userID}
	if parentID == nil {
		q["parent_id"] = nil
	} else {
		q["parent_id"] = *parentID
	}
	cur, err := r.col.Find(ctx, q, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Folder
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *FolderRepo) ListByIDs(ctx context.Context, userID primitive.ObjectID, ids []primitive.ObjectID) ([]domain.Folder, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID, "_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Folder
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *FolderRepo) Rename(ctx context.Context, userID, id primitive.ObjectID, name string) error {
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "user_id": userID},
		bson.M{"$set": bson.M{"name": name, "updated_at": time.Now().UTC()}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FolderRepo) UpdatePathAndParent(ctx context.Context, userID, id primitive.ObjectID, parentID *primitive.ObjectID, newPath string, depth int) error {
	set := bson.M{
		"path":       newPath,
		"depth":      depth,
		"parent_id":  parentID,
		"updated_at": time.Now().UTC(),
	}
	res, err := r.col.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RewriteSubtreePath replaces `oldPrefix` with `newPrefix` in every descendant's `path`
// and adjusts their depth by `depthDelta`. Called after a folder is moved.
func (r *FolderRepo) RewriteSubtreePath(ctx context.Context, userID primitive.ObjectID, oldPrefix, newPrefix string, depthDelta int) error {
	filter := bson.M{
		"user_id": userID,
		"path":    bson.M{"$regex": "^" + regexp.QuoteMeta(oldPrefix)},
	}
	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var f domain.Folder
		if err := cur.Decode(&f); err != nil {
			return err
		}
		newPath := newPrefix + strings.TrimPrefix(f.Path, oldPrefix)
		if _, err := r.col.UpdateByID(ctx, f.ID, bson.M{"$set": bson.M{
			"path":       newPath,
			"depth":      f.Depth + depthDelta,
			"updated_at": time.Now().UTC(),
		}}); err != nil {
			return err
		}
	}
	return cur.Err()
}

func (r *FolderRepo) ListDescendantIDs(ctx context.Context, userID, folderID primitive.ObjectID) ([]primitive.ObjectID, error) {
	f, err := r.Find(ctx, userID, folderID)
	if err != nil {
		return nil, err
	}
	prefix := f.Path + f.ID.Hex() + ","
	cur, err := r.col.Find(ctx, bson.M{
		"user_id": userID,
		"path":    bson.M{"$regex": "^" + regexp.QuoteMeta(prefix)},
	}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var docs []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, 0, len(docs)+1)
	ids = append(ids, folderID)
	for _, d := range docs {
		ids = append(ids, d.ID)
	}
	return ids, nil
}

func (r *FolderRepo) DeleteSubtree(ctx context.Context, userID, folderID primitive.ObjectID) error {
	f, err := r.Find(ctx, userID, folderID)
	if err != nil {
		return err
	}
	prefix := f.Path + f.ID.Hex() + ","
	_, err = r.col.DeleteMany(ctx, bson.M{
		"user_id": userID,
		"$or": []bson.M{
			{"_id": folderID},
			{"path": bson.M{"$regex": "^" + regexp.QuoteMeta(prefix)}},
		},
	})
	return err
}

func (r *FolderRepo) UpdateTagIDs(ctx context.Context, userID, id primitive.ObjectID, add, remove *primitive.ObjectID) error {
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
