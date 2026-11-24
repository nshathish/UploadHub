package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/nshathish/uploadhub/backend/internal/auth"
)

type SessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (auth.User, error)
}

func SessionAuth(authenticator SessionAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		user, err := authenticator.Authenticate(r.Context(), token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ownerContextKey{}, user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
