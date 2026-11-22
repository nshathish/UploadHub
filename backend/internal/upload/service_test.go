package upload

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSigner struct {
	url      string
	err      error
	blobName string
	expires  time.Time
}

type fakeSessionRepository struct {
	created UploadSession
	err     error
}

func (f *fakeSessionRepository) Create(_ context.Context, session UploadSession) error {
	f.created = session
	return f.err
}

func (f *fakeSessionRepository) GetByID(_ context.Context, _ string) (UploadSession, error) {
	return UploadSession{}, ErrSessionNotFound
}

func (f *fakeSigner) GenerateUploadURL(blobName string, expiresAt time.Time) (string, error) {
	f.blobName = blobName
	f.expires = expiresAt
	return f.url, f.err
}

func TestServiceInitiate(t *testing.T) {
	now := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	signer := &fakeSigner{url: "https://storage.invalid/upload"}
	sessions := &fakeSessionRepository{}
	service := NewService(signer, sessions)
	service.now = func() time.Time { return now }
	service.newID = func() (string, error) { return "opaque-id", nil }

	result, err := service.Initiate(context.Background(), "demo-user", "passport.PDF", 1024)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	if result.ID != "opaque-id" || result.BlobName != "opaque-id.pdf" {
		t.Fatalf("Initiate() result = %#v", result)
	}
	if signer.blobName != "opaque-id.pdf" {
		t.Errorf("signed blob = %q", signer.blobName)
	}
	if result.ExpiresAt != now.Add(URLLifetime) {
		t.Errorf("ExpiresAt = %v", result.ExpiresAt)
	}
	if sessions.created.OwnerID != "demo-user" || sessions.created.Status != SessionInitiated {
		t.Errorf("created session = %#v", sessions.created)
	}
}

func TestServiceInitiateReturnsValidationError(t *testing.T) {
	service := NewService(&fakeSigner{}, &fakeSessionRepository{})

	_, err := service.Initiate(context.Background(), "demo-user", "passport.exe", 1024)
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Initiate() error = %v, want %v", err, ErrUnsupportedType)
	}
}
