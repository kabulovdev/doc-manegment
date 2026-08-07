package ports

import (
	"context"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepo interface {
	Create(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	TouchLastLogin(ctx context.Context, id primitive.ObjectID) error
	UpdateDisplayName(ctx context.Context, id primitive.ObjectID, displayName string) error
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, t *domain.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	MarkReplaced(ctx context.Context, id, replacedBy primitive.ObjectID) error
	RevokeFamily(ctx context.Context, familyID primitive.ObjectID) error
}

type StorageConfigRepo interface {
	Create(ctx context.Context, s *domain.StorageConfig) error
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.StorageConfig, error)
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.StorageConfig, error)
	Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error
	IncrementUsage(ctx context.Context, id primitive.ObjectID, deltaBytes, deltaCount int64) error
	SetUsage(ctx context.Context, id primitive.ObjectID, totalBytes, totalCount int64, lastSync time.Time, lastErr string) error
	Delete(ctx context.Context, userID, id primitive.ObjectID) error
}

type FileListFilter struct {
	FolderID   *primitive.ObjectID
	StorageID  *primitive.ObjectID
	TagID      *primitive.ObjectID
	Status     *domain.FileStatus
	Starred    *bool
	Query      string
	FieldKey   string
	FieldValue string
	Limit      int64
	Skip       int64
}

type FileRepo interface {
	Create(ctx context.Context, f *domain.File) error
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.File, error)
	List(ctx context.Context, userID primitive.ObjectID, f FileListFilter) ([]domain.File, error)
	Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error
	RawUpdate(ctx context.Context, userID, id primitive.ObjectID, ops bson.M) error
	Delete(ctx context.Context, userID, id primitive.ObjectID) error
	HardDelete(ctx context.Context, userID, id primitive.ObjectID) error
	HasFilesInStorage(ctx context.Context, storageID primitive.ObjectID) (bool, error)
	SoftDeleteByStorage(ctx context.Context, userID, storageID primitive.ObjectID) error
	UpdateTagIDs(ctx context.Context, userID, id primitive.ObjectID, add, remove *primitive.ObjectID) error
	RemoveTagFromAll(ctx context.Context, userID, tagID primitive.ObjectID) error
	SoftDeleteByFolderIDs(ctx context.Context, userID primitive.ObjectID, folderIDs []primitive.ObjectID) error
	SetStarred(ctx context.Context, userID, id primitive.ObjectID, starred bool) error
	ListDeletedOlderThan(ctx context.Context, before time.Time, limit int64) ([]domain.File, error)
}

type FolderRepo interface {
	Create(ctx context.Context, f *domain.Folder) error
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.Folder, error)
	ListChildren(ctx context.Context, userID primitive.ObjectID, parentID *primitive.ObjectID) ([]domain.Folder, error)
	ListByIDs(ctx context.Context, userID primitive.ObjectID, ids []primitive.ObjectID) ([]domain.Folder, error)
	Rename(ctx context.Context, userID, id primitive.ObjectID, name string) error
	UpdatePathAndParent(ctx context.Context, userID, id primitive.ObjectID, parentID *primitive.ObjectID, newPath string, depth int) error
	RewriteSubtreePath(ctx context.Context, userID primitive.ObjectID, oldPrefix, newPrefix string, depthDelta int) error
	ListDescendantIDs(ctx context.Context, userID, folderID primitive.ObjectID) ([]primitive.ObjectID, error)
	DeleteSubtree(ctx context.Context, userID, folderID primitive.ObjectID) error
	UpdateTagIDs(ctx context.Context, userID, id primitive.ObjectID, add, remove *primitive.ObjectID) error
}

type TagRepo interface {
	Create(ctx context.Context, t *domain.Tag) error
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.Tag, error)
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.Tag, error)
	Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error
	Delete(ctx context.Context, userID, id primitive.ObjectID) error
}

type ShareRepo interface {
	Create(ctx context.Context, s *domain.ShareLink) error
	FindByToken(ctx context.Context, token string) (*domain.ShareLink, error)
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.ShareLink, error)
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.ShareLink, error)
	Revoke(ctx context.Context, userID, id primitive.ObjectID) error
	ConsumeOneTime(ctx context.Context, id primitive.ObjectID) (bool, error)
}

type ShareAccessLogRepo interface {
	Create(ctx context.Context, l *domain.ShareAccessLog) error
	TailByShare(ctx context.Context, shareID primitive.ObjectID, limit int64) ([]domain.ShareAccessLog, error)
}

type ActivityFilter struct {
	UserID      primitive.ObjectID
	SubjectID   *primitive.ObjectID
	SubjectType *domain.ActivitySubjectType
	Limit       int64
}

type ActivityLogRepo interface {
	Create(ctx context.Context, l *domain.ActivityLog) error
	List(ctx context.Context, f ActivityFilter) ([]domain.ActivityLog, error)
}

type APITokenRepo interface {
	Create(ctx context.Context, t *domain.APIToken) error
	FindByHash(ctx context.Context, hash string) (*domain.APIToken, error)
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.APIToken, error)
	Revoke(ctx context.Context, userID, id primitive.ObjectID) error
	TouchLastUsed(ctx context.Context, id primitive.ObjectID, when time.Time) error
}

type AIConfigRepo interface {
	Create(ctx context.Context, c *domain.AIConfig) error
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.AIConfig, error)
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.AIConfig, error)
	Update(ctx context.Context, userID, id primitive.ObjectID, set bson.M) error
	ClearDefault(ctx context.Context, userID primitive.ObjectID, capability domain.AICapability) error
	IncrementUsage(ctx context.Context, id primitive.ObjectID, tokensIn, tokensOut int64, when time.Time) error
	Delete(ctx context.Context, userID, id primitive.ObjectID) error
}

type AIUsageTotals struct {
	TokensIn  int64
	TokensOut int64
	Calls     int64
}

type AIUsageRepo interface {
	Create(ctx context.Context, u *domain.AIUsage) error
	TotalsByConfig(ctx context.Context, userID primitive.ObjectID) (map[primitive.ObjectID]AIUsageTotals, error)
}

type ExtractionRepo interface {
	FindByFile(ctx context.Context, userID, fileID primitive.ObjectID) (*domain.FileExtraction, error)
	Upsert(ctx context.Context, e *domain.FileExtraction) error
}
