package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivitySubjectType string

const (
	ActivitySubjectFile    ActivitySubjectType = "file"
	ActivitySubjectFolder  ActivitySubjectType = "folder"
	ActivitySubjectTag     ActivitySubjectType = "tag"
	ActivitySubjectStorage ActivitySubjectType = "storage"
	ActivitySubjectShare   ActivitySubjectType = "share"
)

// ActivityAction is a free-form verb such as "upload_complete", "delete",
// "restore", "rename", "move", "create", "revoke".
type ActivityAction string

type ActivityLog struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"`
	UserID      primitive.ObjectID  `bson:"user_id"`
	SubjectType ActivitySubjectType `bson:"subject_type"`
	SubjectID   primitive.ObjectID  `bson:"subject_id"`
	Action      ActivityAction      `bson:"action"`
	Metadata    map[string]any      `bson:"metadata,omitempty"`
	IP          string              `bson:"ip,omitempty"`
	UserAgent   string              `bson:"user_agent,omitempty"`
	CreatedAt   time.Time           `bson:"created_at"`
}
