package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Folder struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	UserID    primitive.ObjectID   `bson:"user_id"`
	Name      string               `bson:"name"`
	ParentID  *primitive.ObjectID  `bson:"parent_id,omitempty"`
	Path      string               `bson:"path"` // ",<pid1>,<pid2>," ID-based
	Depth     int                  `bson:"depth"`
	TagIDs    []primitive.ObjectID `bson:"tag_ids"`
	CreatedAt time.Time            `bson:"created_at"`
	UpdatedAt time.Time            `bson:"updated_at"`
}

type Tag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Name      string             `bson:"name"`
	Color     string             `bson:"color"`
	CreatedAt time.Time          `bson:"created_at"`
}

type TagTargetType string

const (
	TagTargetFile   TagTargetType = "file"
	TagTargetFolder TagTargetType = "folder"
)
