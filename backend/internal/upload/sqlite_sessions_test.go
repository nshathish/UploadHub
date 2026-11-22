package upload

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteSessionRepositoryPersistsSession(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "uploadhub.db")
	repository, err := NewSQLiteSessionRepository(databasePath)
	if err != nil {
		t.Fatalf("NewSQLiteSessionRepository() error = %v", err)
	}

	session := validSession()
	if err := repository.Create(context.Background(), session); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repository.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	repository, err = NewSQLiteSessionRepository(databasePath)
	if err != nil {
		t.Fatalf("NewSQLiteSessionRepository() after restart error = %v", err)
	}
	t.Cleanup(func() { _ = repository.Close() })

	stored, err := repository.GetByID(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if stored != session {
		t.Fatalf("GetByID() = %#v, want %#v", stored, session)
	}
}

func TestSQLiteSessionRepositoryReturnsMissingSession(t *testing.T) {
	repository, err := NewSQLiteSessionRepository(filepath.Join(t.TempDir(), "uploadhub.db"))
	if err != nil {
		t.Fatalf("NewSQLiteSessionRepository() error = %v", err)
	}
	t.Cleanup(func() { _ = repository.Close() })

	_, err = repository.GetByID(context.Background(), "missing")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestSQLiteSessionRepositoryRejectsInvalidSession(t *testing.T) {
	repository, err := NewSQLiteSessionRepository(filepath.Join(t.TempDir(), "uploadhub.db"))
	if err != nil {
		t.Fatalf("NewSQLiteSessionRepository() error = %v", err)
	}
	t.Cleanup(func() { _ = repository.Close() })

	err = repository.Create(context.Background(), UploadSession{})
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidSession)
	}
}

func validSession() UploadSession {
	return UploadSession{
		ID:           "opaque-upload-id",
		OwnerID:      "local-demo-user",
		DisplayName:  "synthetic-document.pdf",
		BlobName:     "opaque-upload-id.pdf",
		DeclaredSize: 1024,
		Extension:    ".pdf",
		ContentType:  "application/pdf",
		CreatedAt:    time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC),
		ExpiresAt:    time.Date(2026, time.November, 22, 12, 10, 0, 0, time.UTC),
		Status:       SessionInitiated,
		Version:      1,
	}
}
