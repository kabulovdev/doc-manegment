package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	mongoadapter "gitlab.com/docuemnt_manegment/backend/internal/adapters/mongo"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/crypto"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/qdrant"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/storage/ssrf"
	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/config"
	httpapi "gitlab.com/docuemnt_manegment/backend/internal/http"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"gitlab.com/docuemnt_manegment/backend/internal/services/extractor"
)

type App struct {
	Cfg    *config.Config
	Logger *slog.Logger
	Server *http.Server
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	mc, err := mongoadapter.Connect(ctx, cfg.MongoURI)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	db := mc.Database(cfg.MongoDB)
	if err := mongoadapter.EnsureIndexes(ctx, db); err != nil {
		return nil, fmt.Errorf("ensure indexes: %w", err)
	}

	userRepo := mongoadapter.NewUserRepo(db)
	refreshRepo := mongoadapter.NewRefreshTokenRepo(db)
	storageRepo := mongoadapter.NewStorageConfigRepo(db)
	fileRepo := mongoadapter.NewFileRepo(db)
	folderRepo := mongoadapter.NewFolderRepo(db)
	tagRepo := mongoadapter.NewTagRepo(db)
	shareRepo := mongoadapter.NewShareRepo(db)
	accessLogRepo := mongoadapter.NewShareAccessLogRepo(db)
	activityRepo := mongoadapter.NewActivityLogRepo(db)
	apiTokenRepo := mongoadapter.NewAPITokenRepo(db)
	aiConfigRepo := mongoadapter.NewAIConfigRepo(db)
	aiUsageRepo := mongoadapter.NewAIUsageRepo(db)
	extractionRepo := mongoadapter.NewExtractionRepo(db)

	enc, err := crypto.NewAESGCM(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("init encryptor: %w", err)
	}

	jwtIssuer := auth.NewJWTIssuer(cfg.JWTSecret, cfg.AccessTTL)
	authSvc := services.NewAuthService(userRepo, refreshRepo, jwtIssuer, cfg.RefreshTTL)

	policy := ssrf.DefaultPolicy(cfg.AllowHTTP)
	if cfg.AllowHTTP {
		for _, h := range []string{"minio", "localhost", "127.0.0.1"} {
			policy.AllowedHosts[h] = struct{}{}
		}
	}
	storageSvc := services.NewStorageService(storageRepo, fileRepo, enc, policy)
	fileSvc := services.NewFileService(fileRepo, storageSvc, cfg.MaxFileSize)
	folderSvc := services.NewFolderService(folderRepo, fileRepo)
	tagSvc := services.NewTagService(tagRepo, fileRepo, folderRepo)
	shareSvc := services.NewShareService(shareRepo, accessLogRepo, fileRepo, folderRepo, storageSvc)
	activitySvc := services.NewActivityService(activityRepo)
	apiTokenSvc := services.NewAPITokenService(apiTokenRepo)
	aiSvc := services.NewAIService(aiConfigRepo, aiUsageRepo, enc, policy)
	extractSvc := extractor.NewService(extractionRepo, fileSvc)
	qdrantClient := qdrant.New(cfg.QdrantURL, cfg.QdrantAPIKey)
	embedSvc := services.NewEmbedService(aiSvc, extractSvc, qdrantClient, fileSvc)
	aiFeatures := services.NewAIFeatures(aiSvc, extractSvc, fileSvc, tagRepo, folderRepo, embedSvc)
	aiExtractor := services.NewAIExtractorService(aiSvc, fileSvc, fileRepo, embedSvc)

	router := httpapi.NewRouter(httpapi.Deps{
		Config:             cfg,
		JWT:                jwtIssuer,
		AuthService:        authSvc,
		StorageService:     storageSvc,
		FileService:        fileSvc,
		FolderService:      folderSvc,
		TagService:         tagSvc,
		ShareService:       shareSvc,
		ActivityService:    activitySvc,
		APITokenService:    apiTokenSvc,
		AIService:          aiSvc,
		AIFeatures:         aiFeatures,
		AIExtractorService: aiExtractor,
		EmbedService:       embedSvc,
		FileFinder:         fileRepo,
		FolderFinder:       folderRepo,
	})

	return &App{
		Cfg:    cfg,
		Logger: logger,
		Server: &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: router,
		},
	}, nil
}
