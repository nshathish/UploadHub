package upload

import (
	"context"
	"errors"
	"time"
)

type SessionStatus string

const (
	SessionInitiated  SessionStatus = "initiated"
	SessionUploading  SessionStatus = "uploading"
	SessionUploaded   SessionStatus = "uploaded"
	SessionValidating SessionStatus = "validating"
	SessionAccepted   SessionStatus = "accepted"
	SessionRejected   SessionStatus = "rejected"
	SessionCanceled   SessionStatus = "canceled"
	SessionExpired    SessionStatus = "expired"
	SessionFailed     SessionStatus = "failed"
)

var (
	ErrSessionNotFound        = errors.New("upload session not found")
	ErrInvalidSession         = errors.New("invalid upload session")
	ErrIdempotencyConflict    = errors.New("upload idempotency conflict")
	ErrSessionVersionConflict = errors.New("upload session version conflict")
)

type UploadSession struct {
	ID             string
	OwnerID        string
	DisplayName    string
	BlobName       string
	DeclaredSize   int64
	Extension      string
	ContentType    string
	IdempotencyKey string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	Status         SessionStatus
	Version        int64
}

type SessionRepository interface {
	Create(ctx context.Context, session UploadSession) error
	GetByID(ctx context.Context, id string) (UploadSession, error)
	GetByIdempotencyKey(ctx context.Context, ownerID string, key string) (UploadSession, error)
	UpdateStatus(ctx context.Context, id string, version int64, status SessionStatus) (UploadSession, error)
}

func (s SessionStatus) CanTransitionTo(next SessionStatus) bool {
	allowed := map[SessionStatus][]SessionStatus{
		SessionInitiated:  {SessionUploading, SessionUploaded, SessionRejected, SessionCanceled, SessionExpired, SessionFailed},
		SessionUploading:  {SessionUploaded, SessionCanceled, SessionExpired, SessionFailed},
		SessionUploaded:   {SessionValidating, SessionCanceled, SessionExpired, SessionFailed},
		SessionValidating: {SessionAccepted, SessionRejected, SessionFailed},
	}

	for _, candidate := range allowed[s] {
		if candidate == next {
			return true
		}
	}

	return false
}
