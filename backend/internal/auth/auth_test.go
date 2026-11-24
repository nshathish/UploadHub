package auth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceLoginAuthenticatesAndRevokesSession(t *testing.T) {
	repository, err := NewSQLiteRepository(filepath.Join(t.TempDir(), "uploadhub.db"))
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	service := NewService(repository)
	now := time.Date(2026, time.November, 24, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	user, err := service.CreateUser(context.Background(), "Alex", "UploadHub-Test-2026-Alex!")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.PasswordHash == "UploadHub-Test-2026-Alex!" {
		t.Fatal("password was stored without hashing")
	}

	login, err := service.Login(context.Background(), "alex", "UploadHub-Test-2026-Alex!")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if login.Token == "" || login.ExpiresAt != now.Add(SessionLifetime) {
		t.Fatalf("Login() result = %#v", login)
	}

	authenticatedUser, err := service.Authenticate(context.Background(), login.Token)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if authenticatedUser.ID != user.ID {
		t.Fatalf("Authenticate() user = %#v, want ID %q", authenticatedUser, user.ID)
	}

	if err := service.Logout(context.Background(), login.Token); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	_, err = service.Authenticate(context.Background(), login.Token)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Authenticate() after logout error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestServiceRejectsInvalidCredentials(t *testing.T) {
	repository, err := NewSQLiteRepository(filepath.Join(t.TempDir(), "uploadhub.db"))
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	service := NewService(repository)
	if _, err := service.CreateUser(context.Background(), "morgan", "UploadHub-Test-2026-Morgan!"); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	_, err = service.Login(context.Background(), "morgan", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestSeedDevelopmentUsers(t *testing.T) {
	repository, err := NewSQLiteRepository(filepath.Join(t.TempDir(), "uploadhub.db"))
	if err != nil {
		t.Fatalf("NewSQLiteRepository() error = %v", err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	service := NewService(repository)

	if err := SeedDevelopmentUsers(context.Background(), service); err != nil {
		t.Fatalf("SeedDevelopmentUsers() error = %v", err)
	}
	if err := SeedDevelopmentUsers(context.Background(), service); err != nil {
		t.Fatalf("SeedDevelopmentUsers() repeated error = %v", err)
	}
	for _, testUser := range DevelopmentTestUsers {
		if _, err := service.Login(context.Background(), testUser.Username, testUser.Password); err != nil {
			t.Fatalf("Login(%q) error = %v", testUser.Username, err)
		}
	}
}
