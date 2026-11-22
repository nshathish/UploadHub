package handlers

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
)

type ownerContextKey struct{}

func DevelopmentAuth(token string, ownerID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providedToken, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(providedToken), []byte(token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ownerContextKey{}, ownerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func OwnerID(ctx context.Context) (string, bool) {
	ownerID, ok := ctx.Value(ownerContextKey{}).(string)
	return ownerID, ok
}
