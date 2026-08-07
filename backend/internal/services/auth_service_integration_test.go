//go:build integration

package services_test

import (
	"context"
	"os"
	"testing"
	"time"

	mongoadapter "gitlab.com/docuemnt_manegment/backend/internal/adapters/mongo"
	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Run with: go test -tags=integration ./internal/services/...
// Requires TEST_MONGO_URI pointing to a disposable MongoDB (e.g. mongodb://docmgmt:docmgmtpass@localhost:27018/?authSource=admin).

func requireMongo(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("TEST_MONGO_URI")
	if uri == "" {
		t.Skip("TEST_MONGO_URI not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	dbName := "docmgmt_test_" + time.Now().Format("150405")
	db := client.Database(dbName)
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	require.NoError(t, mongoadapter.EnsureIndexes(ctx, db))
	return db
}

func TestAuthService_RegisterLoginRefresh(t *testing.T) {
	db := requireMongo(t)
	ctx := context.Background()

	userRepo := mongoadapter.NewUserRepo(db)
	refreshRepo := mongoadapter.NewRefreshTokenRepo(db)
	jwt := auth.NewJWTIssuer("test-secret-minimum-32-chars-abcd", 15*time.Minute)
	svc := services.NewAuthService(userRepo, refreshRepo, jwt, 24*time.Hour)

	u, err := svc.Register(ctx, "Alice@example.com ", "hunter2-strong", "Alice")
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", u.Email)

	_, err = svc.Register(ctx, "alice@example.com", "hunter2-strong", "")
	require.ErrorIs(t, err, domain.ErrConflict)

	pair, err := svc.Login(ctx, "alice@example.com", "hunter2-strong", "test-ua", "127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)

	pair2, err := svc.Refresh(ctx, pair.RefreshToken, "test-ua", "127.0.0.1")
	require.NoError(t, err)
	require.NotEqual(t, pair.RefreshToken, pair2.RefreshToken, "refresh token should rotate")

	_, err = svc.Refresh(ctx, pair.RefreshToken, "test-ua", "127.0.0.1")
	require.ErrorIs(t, err, domain.ErrTokenReused, "reuse of old refresh token should revoke the family")

	_, err = svc.Refresh(ctx, pair2.RefreshToken, "test-ua", "127.0.0.1")
	require.ErrorIs(t, err, domain.ErrTokenRevoked, "after reuse detection, new token in family is revoked")
}

func TestAuthService_UpdateProfile(t *testing.T) {
	db := requireMongo(t)
	ctx := context.Background()

	userRepo := mongoadapter.NewUserRepo(db)
	refreshRepo := mongoadapter.NewRefreshTokenRepo(db)
	jwt := auth.NewJWTIssuer("test-secret-minimum-32-chars-abcd", 15*time.Minute)
	svc := services.NewAuthService(userRepo, refreshRepo, jwt, 24*time.Hour)

	u, err := svc.Register(ctx, "carol@example.com", "hunter2-strong", "Carol")
	require.NoError(t, err)
	require.Equal(t, "Carol", u.DisplayName)

	updated, err := svc.UpdateProfile(ctx, u.ID, "  Carol Updated  ")
	require.NoError(t, err)
	require.Equal(t, "Carol Updated", updated.DisplayName)

	me, err := svc.Me(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, "Carol Updated", me.DisplayName)

	_, err = svc.UpdateProfile(ctx, u.ID, "")
	require.ErrorIs(t, err, domain.ErrInvalidInput)

	_, err = svc.UpdateProfile(ctx, u.ID, stringOfLen(101))
	require.ErrorIs(t, err, domain.ErrInvalidInput)
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

func TestAuthService_InvalidCreds(t *testing.T) {
	db := requireMongo(t)
	ctx := context.Background()

	userRepo := mongoadapter.NewUserRepo(db)
	refreshRepo := mongoadapter.NewRefreshTokenRepo(db)
	jwt := auth.NewJWTIssuer("test-secret-minimum-32-chars-abcd", 15*time.Minute)
	svc := services.NewAuthService(userRepo, refreshRepo, jwt, 24*time.Hour)

	_, err := svc.Register(ctx, "bob@example.com", "hunter2-strong", "Bob")
	require.NoError(t, err)

	_, err = svc.Login(ctx, "bob@example.com", "wrong", "ua", "ip")
	require.ErrorIs(t, err, domain.ErrInvalidCreds)

	_, err = svc.Login(ctx, "nobody@example.com", "hunter2-strong", "ua", "ip")
	require.ErrorIs(t, err, domain.ErrInvalidCreds)
}

// Silence unused import when build tag excludes this file.
var _ = bson.M{}
