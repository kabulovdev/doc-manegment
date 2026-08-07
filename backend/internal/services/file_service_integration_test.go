//go:build integration

package services_test

import (
	"context"
	"testing"
	"time"

	mongoadapter "gitlab.com/docuemnt_manegment/backend/internal/adapters/mongo"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFileRepo_TrashWorkflow verifies the soft-delete → list → restore → list
// cycle at the repository layer. It doesn't exercise the bucket object flow
// (which requires a real S3-compatible endpoint); that is covered in manual
// end-to-end smoke tests.
func TestFileRepo_TrashWorkflow(t *testing.T) {
	db := requireMongo(t)
	ctx := context.Background()

	repo := mongoadapter.NewFileRepo(db)
	userID := primitive.NewObjectID()
	storageID := primitive.NewObjectID()

	f := &domain.File{
		UserID:    userID,
		StorageID: storageID,
		Name:      "doc.pdf",
		ObjectKey: "u/" + userID.Hex() + "/doc.pdf",
		SizeBytes: 1024,
		MimeType:  "application/pdf",
		Status:    domain.FileStatusReady,
		TagIDs:    []primitive.ObjectID{},
	}
	require.NoError(t, repo.Create(ctx, f))

	// Soft-delete.
	require.NoError(t, repo.Delete(ctx, userID, f.ID))
	got, err := repo.Find(ctx, userID, f.ID)
	require.NoError(t, err)
	require.Equal(t, domain.FileStatusDeleted, got.Status)

	// Default list filter excludes deleted.
	list, err := repo.List(ctx, userID, ports.FileListFilter{})
	require.NoError(t, err)
	for _, x := range list {
		require.NotEqual(t, f.ID, x.ID)
	}

	// status=deleted returns it.
	del := domain.FileStatusDeleted
	list, err = repo.List(ctx, userID, ports.FileListFilter{Status: &del})
	require.NoError(t, err)
	var found bool
	for _, x := range list {
		if x.ID == f.ID {
			found = true
		}
	}
	require.True(t, found, "expected soft-deleted file in status=deleted listing")

	// Restore.
	require.NoError(t, repo.Update(ctx, userID, f.ID, map[string]any{"status": domain.FileStatusReady}))
	got, err = repo.Find(ctx, userID, f.ID)
	require.NoError(t, err)
	require.Equal(t, domain.FileStatusReady, got.Status)

	// ListDeletedOlderThan is empty for future-cutoff; older cutoff returns [].
	old, err := repo.ListDeletedOlderThan(ctx, time.Now().Add(-time.Hour), 10)
	require.NoError(t, err)
	require.Empty(t, old)

	// Soft-delete again and backdate updated_at to trigger purge selection.
	require.NoError(t, repo.Delete(ctx, userID, f.ID))
	require.NoError(t, repo.Update(ctx, userID, f.ID, map[string]any{
		"updated_at": time.Now().UTC().Add(-48 * time.Hour),
	}))
	old, err = repo.ListDeletedOlderThan(ctx, time.Now().Add(-24*time.Hour), 10)
	require.NoError(t, err)
	require.NotEmpty(t, old)

	// Hard delete removes the doc.
	require.NoError(t, repo.HardDelete(ctx, userID, f.ID))
	_, err = repo.Find(ctx, userID, f.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
