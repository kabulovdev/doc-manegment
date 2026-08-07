package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// APIToken is a personal access token that lets external tools authenticate
// against the API instead of using a JWT session.
//
// Storage: only the SHA-256 hash of the plaintext is kept; the plaintext is
// returned exactly once on creation. Prefix is the first chunk of the secret
// (after the "doc_" marker) and is safe to display for identification.
type APIToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id"`
	Name       string             `bson:"name"`
	Hash       string             `bson:"hash"`
	Prefix     string             `bson:"prefix"`
	Scopes     []string           `bson:"scopes"`
	CreatedAt  time.Time          `bson:"created_at"`
	LastUsedAt *time.Time         `bson:"last_used_at,omitempty"`
	RevokedAt  *time.Time         `bson:"revoked_at,omitempty"`
}

// APITokenPrefix is the plaintext prefix that marks API tokens apart from JWTs.
const APITokenPrefix = "doc_"
