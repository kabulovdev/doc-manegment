// Purge is a standalone job that hard-deletes soft-deleted files whose trash
// retention window has elapsed. It removes the bucket object (via the per-file
// storage provider) and the Mongo document.
//
// Intended to be scheduled (e.g. once a day via cron or a k8s CronJob):
//
//	MONGO_URI=... MASTER_KEY=... JWT_SECRET=... ./purge --retention=720h
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	mongoadapter "gitlab.com/docuemnt_manegment/backend/internal/adapters/mongo"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/crypto"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/storage/ssrf"
	"gitlab.com/docuemnt_manegment/backend/internal/config"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
)

const defaultRetention = 30 * 24 * time.Hour

func main() {
	retention := flag.Duration("retention", defaultRetention, "minimum age (e.g. 720h) of soft-deleted files before hard deletion")
	batch := flag.Int64("batch", 200, "max files per batch")
	dryRun := flag.Bool("dry-run", false, "log files that would be purged but don't touch storage")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	mc, err := mongoadapter.Connect(ctx, cfg.MongoURI)
	if err != nil {
		slog.Error("mongo connect", "err", err)
		os.Exit(1)
	}
	defer func() { _ = mc.Disconnect(context.Background()) }()
	db := mc.Database(cfg.MongoDB)

	fileRepo := mongoadapter.NewFileRepo(db)
	storageRepo := mongoadapter.NewStorageConfigRepo(db)

	enc, err := crypto.NewAESGCM(cfg.MasterKey)
	if err != nil {
		slog.Error("init encryptor", "err", err)
		os.Exit(1)
	}

	policy := ssrf.DefaultPolicy(cfg.AllowHTTP)
	if cfg.AllowHTTP {
		for _, h := range []string{"minio", "localhost", "127.0.0.1"} {
			policy.AllowedHosts[h] = struct{}{}
		}
	}
	storageSvc := services.NewStorageService(storageRepo, fileRepo, enc, policy)
	fileSvc := services.NewFileService(fileRepo, storageSvc, cfg.MaxFileSize)

	before := time.Now().UTC().Add(-*retention)
	slog.Info("purge starting", "before", before, "retention", retention.String(), "dry_run", *dryRun)

	if *dryRun {
		list, err := fileRepo.ListDeletedOlderThan(ctx, before, *batch)
		if err != nil {
			slog.Error("list deleted", "err", err)
			os.Exit(1)
		}
		for _, f := range list {
			slog.Info("would purge", "file_id", f.ID.Hex(), "user_id", f.UserID.Hex(), "name", f.Name, "object_key", f.ObjectKey, "updated_at", f.UpdatedAt)
		}
		slog.Info("purge dry-run complete", "count", len(list))
		return
	}

	purged, errs := fileSvc.PurgeOlderThan(ctx, before, *batch)
	for _, e := range errs {
		slog.Warn("purge error", "err", e)
	}
	slog.Info("purge complete", "purged", purged, "errors", len(errs))
	if len(errs) > 0 {
		os.Exit(2)
	}
}
