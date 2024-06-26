package middleware

import (
	"avito-test-task/internal/lib/api"
	"avito-test-task/internal/lib/api/jsn"
	"avito-test-task/internal/lib/api/msg"
	"avito-test-task/internal/lib/jwt"
	"avito-test-task/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strings"
)

const Authorization = "Authorization"

// NewAuthorizationMiddleware creates a new authorization middleware.
// It checks the Authorization header for a valid JWT token.
// If the token is valid, it extracts the role from it and adds it to the request context.
func NewAuthorizationMiddleware(logger *slog.Logger, manager *jwt.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(Authorization)
			if token == "" {
				logger.Error("nothing in Authorization header")
				jsn.EncodeResponse(w, http.StatusUnauthorized, api.ErrResponse(msg.APINotAuthorized), logger)
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")

			err := manager.VerifyToken(token)
			if err != nil {
				logger.Error("invalid jwt token", sl.Err(err))
				jsn.EncodeResponse(w, http.StatusUnauthorized, api.ErrResponse(msg.APINotAuthorized), logger)
				return
			}

			role, err := manager.GetRole(token)
			if err != nil {
				logger.Error("failed to get role from token", sl.Err(err))
				jsn.EncodeResponse(w, http.StatusUnauthorized, api.ErrResponse(msg.APINotAuthorized), logger)
				return
			}

			r = api.SetUserRole(r, role)

			next.ServeHTTP(w, r)
		})
	}
}

func EnsureAdmin(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := api.UserRole(r)
		if role != "admin" {
			jsn.EncodeResponse(w, http.StatusForbidden, api.ErrResponse(msg.APIForbidden), logger)
			return
		}

		next.ServeHTTP(w, r)
	})
}
