package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"password_hash"`
	DisplayName  string             `bson:"display_name"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
	LastLoginAt  *time.Time         `bson:"last_login_at,omitempty"`
}

type RefreshToken struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty"`
	UserID     primitive.ObjectID  `bson:"user_id"`
	TokenHash  string              `bson:"token_hash"`
	FamilyID   primitive.ObjectID  `bson:"family_id"`
	ReplacedBy *primitive.ObjectID `bson:"replaced_by,omitempty"`
	Revoked    bool                `bson:"revoked"`
	ExpiresAt  time.Time           `bson:"expires_at"`
	CreatedAt  time.Time           `bson:"created_at"`
	UserAgent  string              `bson:"user_agent,omitempty"`
	IP         string              `bson:"ip,omitempty"`
}
