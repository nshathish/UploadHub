package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/nshathish/uploadhub/backend/internal/auth"
)

type LoginService interface {
	Login(ctx context.Context, username string, password string) (auth.LoginResult, error)
}

type LogoutService interface {
	Logout(ctx context.Context, token string) error
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      LoginUser `json:"user"`
}

type LoginUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewLoginHandler(service LoginService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request LoginRequest
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "invalid JSON request", http.StatusBadRequest)
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			http.Error(w, "unexpected request content", http.StatusBadRequest)
			return
		}
		result, err := service.Login(r.Context(), request.Username, request.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				http.Error(w, "invalid username or password", http.StatusUnauthorized)
				return
			}
			http.Error(w, "unable to sign in", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, LoginResponse{Token: result.Token, ExpiresAt: result.ExpiresAt, User: LoginUser{ID: result.User.ID, Username: result.User.Username}})
	}
}

func NewLogoutHandler(service LogoutService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := service.Logout(r.Context(), token); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
