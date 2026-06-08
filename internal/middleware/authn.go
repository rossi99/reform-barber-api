package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/reform-barber/api/internal/auth"
	"github.com/reform-barber/api/internal/respond"
)

type contextKey string

const claimsKey contextKey = "claims"

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				respond.Error(w, http.StatusUnauthorized, "missing or malformed token")
				return
			}
			claims, err := auth.ParseAccessToken(jwtSecret, strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromCtx(ctx context.Context) *auth.Claims {
	c, _ := ctx.Value(claimsKey).(*auth.Claims)
	return c
}

func UserIDFromCtx(ctx context.Context) uuid.UUID {
	c := ClaimsFromCtx(ctx)
	if c == nil {
		return uuid.Nil
	}
	id, _ := uuid.Parse(c.Subject)
	return id
}
