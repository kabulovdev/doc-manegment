package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/config"
	"gitlab.com/docuemnt_manegment/backend/internal/http/handlers"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
)

type Deps struct {
	Config             *config.Config
	JWT                *auth.JWTIssuer
	AuthService        *services.AuthService
	StorageService     *services.StorageService
	FileService        *services.FileService
	FolderService      *services.FolderService
	TagService         *services.TagService
	ShareService       *services.ShareService
	ActivityService    *services.ActivityService
	APITokenService    *services.APITokenService
	AIService          *services.AIService
	AIFeatures         *services.AIFeatures
	AIExtractorService *services.AIExtractorService
	EmbedService       *services.EmbedService
	FileFinder         services.FileFinder
	FolderFinder       services.FolderFinder
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(10 * time.Minute))
	r.Use(mw.CORS(d.Config.FrontendOrigin))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	v := validator.New(validator.WithRequiredStructEnabled())
	authH := handlers.NewAuthHandler(d.AuthService, v, d.Config)
	storageH := handlers.NewStorageHandler(d.StorageService, v)
	fileH := handlers.NewFileHandler(d.FileService, d.ActivityService, d.EmbedService, v)
	folderH := handlers.NewFolderHandler(d.FolderService, d.ActivityService, v)
	tagH := handlers.NewTagHandler(d.TagService, v)
	shareH := handlers.NewShareHandler(d.ShareService, d.ActivityService, v)
	activityH := handlers.NewActivityHandler(d.ActivityService)
	tokenH := handlers.NewAPITokenHandler(d.APITokenService, v)
	aiH := handlers.NewAIHandler(d.AIService, d.AIFeatures, d.AIExtractorService, v)
	publicShareH := handlers.NewPublicShareHandler(d.ShareService, d.FileFinder, d.FolderFinder, v, d.Config)

	shareLimiter := mw.NewRateLimiter(120, mw.KeyByIP).Middleware()
	authLimiter := mw.NewRateLimiter(30, mw.KeyByIP).Middleware()

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(authLimiter)
				r.Post("/register", authH.Register)
				r.Post("/login", authH.Login)
			})
			r.Post("/refresh", authH.Refresh)
			r.Post("/logout", authH.Logout)
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireAuth(d.JWT, d.APITokenService))
				r.Get("/me", authH.Me)
				r.Patch("/me", authH.UpdateMe)

				r.Get("/tokens", tokenH.List)
				r.Post("/tokens", tokenH.Create)
				r.Delete("/tokens/{id}", tokenH.Delete)
				r.Get("/token-scopes", tokenH.Scopes)
			})
		})

		// Public share routes: no auth, rate-limited.
		r.Route("/share/{token}", func(r chi.Router) {
			r.Use(shareLimiter)
			r.Get("/", publicShareH.Metadata)
			r.Post("/unlock", publicShareH.Unlock)
			r.Get("/content", publicShareH.Content)
			r.Get("/children", publicShareH.FolderChildren)
		})

		r.Group(func(r chi.Router) {
			r.Use(mw.RequireAuth(d.JWT, d.APITokenService))

			r.Route("/storages", func(r chi.Router) {
				r.Get("/", storageH.List)
				r.Post("/", storageH.Create)
				r.Get("/{id}", storageH.Get)
				r.Delete("/{id}", storageH.Delete)
				r.Post("/{id}/test", storageH.Test)
				r.Post("/{id}/resync", storageH.Resync)
			})

			r.Route("/folders", func(r chi.Router) {
				r.Get("/", folderH.List)
				r.Post("/", folderH.Create)
				r.Get("/{id}", folderH.Get)
				r.Patch("/{id}", folderH.Rename)
				r.Post("/{id}/move", folderH.Move)
				r.Delete("/{id}", folderH.Delete)
			})

			r.Route("/tags", func(r chi.Router) {
				r.Get("/", tagH.List)
				r.Post("/", tagH.Create)
				r.Patch("/{id}", tagH.Update)
				r.Delete("/{id}", tagH.Delete)
				r.Post("/{id}/attach", tagH.Attach)
				r.Post("/{id}/detach", tagH.Detach)
			})

			r.Route("/files", func(r chi.Router) {
				r.Get("/", fileH.List)
				r.Post("/upload/init", fileH.InitUpload)
				r.Put("/upload/{id}/part", fileH.UploadPart)
				r.Post("/upload/{id}/complete", fileH.Complete)
				r.Post("/upload/{id}/abort", fileH.Abort)
				r.Post("/upload/{id}/single", fileH.UploadSingle)
				r.Get("/{id}", fileH.Get)
				r.Patch("/{id}", fileH.Update)
				r.Delete("/{id}", fileH.Delete)
				r.Post("/{id}/star", fileH.Star)
				r.Post("/{id}/restore", fileH.Restore)
				r.Post("/{id}/purge", fileH.Purge)
				r.Get("/{id}/content", fileH.Stream)
				r.Post("/{id}/ai/extract-fields", aiH.SuggestFields)
				r.Post("/{id}/ai/suggest-tags", aiH.SuggestTags)
				r.Post("/{id}/ai/suggest-name", aiH.SuggestName)
				r.Post("/{id}/ai/suggest-folder", aiH.SuggestFolder)
				r.Post("/{id}/ai/process", aiH.ProcessDocument)
				r.Get("/{id}/ai/extraction", aiH.GetExtraction)
				r.Post("/{id}/ai/chat", aiH.FileChat)
			})

			r.Route("/shares", func(r chi.Router) {
				r.Get("/", shareH.List)
				r.Post("/", shareH.Create)
				r.Get("/{id}", shareH.Get)
				r.Delete("/{id}", shareH.Delete)
			})

			r.Route("/activity", func(r chi.Router) {
				r.Get("/", activityH.List)
				r.Get("/recent", activityH.Recent)
			})

			r.Route("/ai", func(r chi.Router) {
				r.Route("/configs", func(r chi.Router) {
					r.Get("/", aiH.List)
					r.Post("/", aiH.Create)
					r.Get("/{id}", aiH.Get)
					r.Patch("/{id}", aiH.Update)
					r.Delete("/{id}", aiH.Delete)
					r.Post("/{id}/test", aiH.Test)
					r.Post("/{id}/default", aiH.SetDefault)
				})
				r.Get("/usage", aiH.Usage)
				r.Post("/search", aiH.Search)
				r.Post("/ask", aiH.Ask)
			})
		})

		// Load/autoscaling test endpoint. Disabled unless ENABLE_LOAD_TEST=true.
		if d.Config.EnableLoadTest {
			debugH := handlers.NewDebugHandler()
			r.Route("/debug", func(r chi.Router) {
				r.Get("/cpu-load", debugH.CPULoad)
				r.Post("/cpu-load", debugH.CPULoad)
			})
		}
	})

	return r
}
