package util

import (
	"crypto/rand"
	"encoding/base64"
	"path"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectKey builds an S3 object key that isolates each user under a prefix and
// adds a random suffix so filename collisions cannot overwrite prior objects.
// Format: {userID}/{randomID}/{sanitized-name}
func ObjectKey(userID primitive.ObjectID, name string) string {
	b := make([]byte, 9)
	_, _ = rand.Read(b)
	rand := base64.RawURLEncoding.EncodeToString(b)
	clean := path.Base(strings.ReplaceAll(strings.TrimSpace(name), "\\", "/"))
	if clean == "" || clean == "." || clean == "/" {
		clean = "file"
	}
	return userID.Hex() + "/" + rand + "/" + clean
}
