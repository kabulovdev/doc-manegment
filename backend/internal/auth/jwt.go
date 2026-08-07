package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidToken = errors.New("invalid token")

type JWTIssuer struct {
	secret    []byte
	accessTTL time.Duration
}

func NewJWTIssuer(secret string, accessTTL time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), accessTTL: accessTTL}
}

func (j *JWTIssuer) IssueAccess(userID primitive.ObjectID) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(j.secret)
}

func (j *JWTIssuer) Parse(tokenStr string) (*Claims, error) {
	c := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return c, nil
}

func NewRefreshToken() (plaintext string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	plaintext = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(plaintext))
	hash = base64.RawStdEncoding.EncodeToString(sum[:])
	return plaintext, hash, nil
}

func HashRefreshToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

// NewAPIToken produces a random personal-access token of the form
// `doc_<44 base64url chars>`. The plaintext is returned once; only the
// SHA-256 hash is persisted. The prefix (first 8 chars after `doc_`) is safe
// to expose for identification and is returned alongside.
func NewAPIToken() (plaintext, hash, prefix string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(b)
	plaintext = "doc_" + secret
	sum := sha256.Sum256([]byte(plaintext))
	hash = base64.RawStdEncoding.EncodeToString(sum[:])
	if len(secret) >= 8 {
		prefix = secret[:8]
	} else {
		prefix = secret
	}
	return plaintext, hash, prefix, nil
}

// HashAPIToken returns the stored hash for a given plaintext token. Used by the
// auth middleware to look up `doc_*` bearers.
func HashAPIToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}
