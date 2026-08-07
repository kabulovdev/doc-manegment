package middleware

import (
	"context"
	"net/http"
	"strings"

	"gitlab.com/docuemnt_manegment/backend/internal/auth"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ctxKey string

const (
	userIDKey ctxKey = "uid"
	scopesKey ctxKey = "scopes"
)

// RequireAuth accepts either a JWT session token or a personal-access API
// token (prefixed with `doc_`) in the `Authorization: Bearer …` header. When
// apiTokens is nil, only JWT auth is accepted.
func RequireAuth(j *auth.JWTIssuer, apiTokens *services.APITokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				respond.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			raw := strings.TrimPrefix(h, "Bearer ")

			if apiTokens != nil && strings.HasPrefix(raw, domain.APITokenPrefix) {
				t, err := apiTokens.VerifyBearer(r.Context(), raw)
				if err != nil {
					respond.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
					return
				}
				ctx := context.WithValue(r.Context(), userIDKey, t.UserID)
				ctx = context.WithValue(ctx, scopesKey, t.Scopes)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			claims, err := j.Parse(raw)
			if err != nil {
				respond.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			id, err := primitive.ObjectIDFromHex(claims.UserID)
			if err != nil {
				respond.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (primitive.ObjectID, bool) {
	id, ok := ctx.Value(userIDKey).(primitive.ObjectID)
	return id, ok
}

// ScopesFromContext returns the API-token scopes for the current request if
// the caller authenticated with an API token. Returns nil when authed via
// JWT (session has full access).
func ScopesFromContext(ctx context.Context) []string {
	s, _ := ctx.Value(scopesKey).([]string)
	return s
}
