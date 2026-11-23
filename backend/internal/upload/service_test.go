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
	created     UploadSession
	existing    UploadSession
	err         error
	createCalls int
	byID        UploadSession
	getByIDErr  error
}

type fakeBlobInspector struct {
	properties BlobProperties
	err        error
	blobName   string
}

func (f *fakeSessionRepository) Create(_ context.Context, session UploadSession) error {
	f.createCalls++
	f.created = session
	return f.err
}

func (f *fakeSessionRepository) GetByID(_ context.Context, _ string) (UploadSession, error) {
	if f.getByIDErr != nil {
		return UploadSession{}, f.getByIDErr
	}
	if f.byID.ID == "" {
		return UploadSession{}, ErrSessionNotFound
	}
	return f.byID, nil
}

func (f *fakeSessionRepository) GetByIdempotencyKey(_ context.Context, _ string, _ string) (UploadSession, error) {
	if f.err != nil {
		return UploadSession{}, f.err
	}
	if f.existing.ID != "" {
		return f.existing, nil
	}
	return UploadSession{}, ErrSessionNotFound
}

func (f *fakeSessionRepository) UpdateStatus(_ context.Context, _ string, version int64, status SessionStatus) (UploadSession, error) {
	if f.byID.Version != version {
		return UploadSession{}, ErrSessionVersionConflict
	}
	f.byID.Status = status
	f.byID.Version++
	return f.byID, nil
}

func (f *fakeSigner) GenerateUploadURL(blobName string, expiresAt time.Time) (string, error) {
	f.blobName = blobName
	f.expires = expiresAt
	return f.url, f.err
}

func (f *fakeSigner) GenerateBlockUploadURL(blobName string, expiresAt time.Time) (string, error) {
	f.blobName = blobName
	f.expires = expiresAt
	return f.url, f.err
}

func (f *fakeBlobInspector) GetBlobProperties(_ context.Context, blobName string) (BlobProperties, error) {
	f.blobName = blobName
	return f.properties, f.err
}

func TestServiceInitiate(t *testing.T) {
	now := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	signer := &fakeSigner{url: "https://storage.invalid/upload"}
	sessions := &fakeSessionRepository{}
	service := NewService(signer, &fakeBlobInspector{}, sessions)
	service.now = func() time.Time { return now }
	service.newID = func() (string, error) { return "opaque-id", nil }

	result, err := service.Initiate(context.Background(), "demo-user", "test-key", "passport.PDF", 1024)
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
	service := NewService(&fakeSigner{}, &fakeBlobInspector{}, &fakeSessionRepository{})

	_, err := service.Initiate(context.Background(), "demo-user", "test-key", "passport.exe", 1024)
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("Initiate() error = %v, want %v", err, ErrUnsupportedType)
	}
}

func TestServiceInitiateReusesActiveSession(t *testing.T) {
	now := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	signer := &fakeSigner{url: "https://storage.invalid/upload"}
	sessions := &fakeSessionRepository{existing: UploadSession{
		ID:         "existing-id",
		BlobName:   "existing-id.pdf",
		ExpiresAt:  now.Add(URLLifetime),
		Status:     SessionInitiated,
		UploadMode: UploadModeSingle,
	}}
	service := NewService(signer, &fakeBlobInspector{}, sessions)
	service.now = func() time.Time { return now }

	result, err := service.Initiate(context.Background(), "demo-user", "test-key", "passport.pdf", 1024)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}
	if result.ID != "existing-id" || result.BlobName != "existing-id.pdf" {
		t.Fatalf("Initiate() result = %#v", result)
	}
	if sessions.createCalls != 0 {
		t.Fatalf("Create() calls = %d, want 0", sessions.createCalls)
	}
}

func TestServiceInitiateChunked(t *testing.T) {
	now := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	signer := &fakeSigner{url: "https://storage.invalid/upload"}
	sessions := &fakeSessionRepository{}
	service := NewService(signer, &fakeBlobInspector{}, sessions)
	service.now = func() time.Time { return now }
	service.newID = func() (string, error) { return "opaque-id", nil }

	result, err := service.InitiateChunked(context.Background(), "demo-user", "chunk-key", "passport.pdf", 1024)
	if err != nil {
		t.Fatalf("InitiateChunked() error = %v", err)
	}
	if result.UploadMode != UploadModeChunked || result.BlockSize != ChunkBlockSize {
		t.Fatalf("InitiateChunked() result = %#v", result)
	}
	if sessions.created.UploadMode != UploadModeChunked {
		t.Fatalf("stored mode = %q, want %q", sessions.created.UploadMode, UploadModeChunked)
	}
	if sessions.created.BlockSize != ChunkBlockSize {
		t.Fatalf("stored block size = %d, want %d", sessions.created.BlockSize, ChunkBlockSize)
	}
}

func TestServiceInitiateChunkedAcceptsLargeFiles(t *testing.T) {
	signer := &fakeSigner{url: "https://storage.invalid/upload"}
	sessions := &fakeSessionRepository{}
	service := NewService(signer, &fakeBlobInspector{}, sessions)
	service.newID = func() (string, error) { return "large-id", nil }

	result, err := service.InitiateChunked(context.Background(), "demo-user", "large-key", "large.pdf", MaxSingleFileSize+1)
	if err != nil {
		t.Fatalf("InitiateChunked() error = %v", err)
	}
	if result.BlockSize != ChunkBlockSize || sessions.created.DeclaredSize != MaxSingleFileSize+1 {
		t.Fatalf("InitiateChunked() result = %#v, session = %#v", result, sessions.created)
	}
}

func TestServiceGetRejectsAnotherOwner(t *testing.T) {
	service := NewService(&fakeSigner{}, &fakeBlobInspector{}, &fakeSessionRepository{byID: UploadSession{ID: "upload-id", OwnerID: "other-user"}})

	_, err := service.Get(context.Background(), "demo-user", "upload-id")
	if !errors.Is(err, ErrSessionForbidden) {
		t.Fatalf("Get() error = %v, want %v", err, ErrSessionForbidden)
	}
}

func TestServiceCompleteMarksMatchingBlobUploaded(t *testing.T) {
	now := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	session := validSession()
	session.ExpiresAt = now.Add(URLLifetime)
	sessions := &fakeSessionRepository{byID: session}
	inspector := &fakeBlobInspector{properties: BlobProperties{Size: session.DeclaredSize}}
	service := NewService(&fakeSigner{}, inspector, sessions)
	service.now = func() time.Time { return now }

	result, err := service.Complete(context.Background(), session.OwnerID, session.ID)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if result.Status != SessionUploaded {
		t.Fatalf("status = %q, want %q", result.Status, SessionUploaded)
	}
	if inspector.blobName != session.BlobName {
		t.Fatalf("inspected blob = %q, want %q", inspector.blobName, session.BlobName)
	}
}

func TestServiceCompleteRejectsMissingOrMismatchedBlob(t *testing.T) {
	now := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		inspector fakeBlobInspector
		wantErr   error
	}{
		{name: "missing blob", inspector: fakeBlobInspector{err: ErrBlobNotFound}, wantErr: ErrBlobNotFound},
		{name: "wrong size", inspector: fakeBlobInspector{properties: BlobProperties{Size: 1}}, wantErr: ErrBlobSizeMismatch},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := validSession()
			session.ExpiresAt = now.Add(URLLifetime)
			sessions := &fakeSessionRepository{byID: session}
			service := NewService(&fakeSigner{}, &test.inspector, sessions)
			service.now = func() time.Time { return now }

			_, err := service.Complete(context.Background(), session.OwnerID, session.ID)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Complete() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == ErrBlobSizeMismatch && sessions.byID.Status != SessionRejected {
				t.Fatalf("status = %q, want %q", sessions.byID.Status, SessionRejected)
			}
		})
	}
}
