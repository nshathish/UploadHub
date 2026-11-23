package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	URLLifetime    = 10 * time.Minute
	ChunkBlockSize = 4 << 20
)

var (
	ErrGenerateUploadURL      = errors.New("generate upload URL")
	ErrGenerateID             = errors.New("generate upload ID")
	ErrMissingOwner           = errors.New("missing upload owner")
	ErrMissingRepository      = errors.New("missing upload session repository")
	ErrMissingIdempotencyKey  = errors.New("missing idempotency key")
	ErrCreateSession          = errors.New("create upload session")
	ErrGetIdempotentSession   = errors.New("get idempotent upload session")
	ErrSessionForbidden       = errors.New("upload session access denied")
	ErrMissingBlobInspector   = errors.New("missing blob inspector")
	ErrBlobNotFound           = errors.New("uploaded blob not found")
	ErrBlobSizeMismatch       = errors.New("uploaded blob size does not match")
	ErrSessionExpired         = errors.New("upload session expired")
	ErrInvalidCompletionState = errors.New("invalid upload completion state")
)

type URLSigner interface {
	GenerateUploadURL(blobName string, expiresAt time.Time) (string, error)
	GenerateBlockUploadURL(blobName string, expiresAt time.Time) (string, error)
}

type BlobProperties struct {
	Size        int64
	ContentType string
	ETag        string
}

type BlobInspector interface {
	GetBlobProperties(ctx context.Context, blobName string) (BlobProperties, error)
}

type InitiateResult struct {
	ID         string     `json:"id"`
	BlobName   string     `json:"blobName"`
	UploadURL  string     `json:"uploadUrl"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	UploadMode UploadMode `json:"uploadMode"`
	BlockSize  int64      `json:"blockSize,omitempty"`
}

type SessionResult struct {
	ID        string        `json:"id"`
	FileName  string        `json:"fileName"`
	Size      int64         `json:"size"`
	Status    SessionStatus `json:"status"`
	ExpiresAt time.Time     `json:"expiresAt"`
}

type Service struct {
	signer    URLSigner
	inspector BlobInspector
	sessions  SessionRepository
	now       func() time.Time
	newID     func() (string, error)
}

func NewService(signer URLSigner, inspector BlobInspector, sessions SessionRepository) *Service {
	return &Service{
		signer:    signer,
		inspector: inspector,
		sessions:  sessions,
		now:       time.Now,
		newID:     newID,
	}
}

func (s *Service) Initiate(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64) (InitiateResult, error) {
	return s.initiate(ctx, ownerID, idempotencyKey, fileName, size, UploadModeSingle)
}

func (s *Service) InitiateChunked(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64) (InitiateResult, error) {
	return s.initiate(ctx, ownerID, idempotencyKey, fileName, size, UploadModeChunked)
}

func (s *Service) initiate(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64, uploadMode UploadMode) (InitiateResult, error) {
	if ownerID == "" {
		return InitiateResult{}, ErrMissingOwner
	}
	if s.sessions == nil {
		return InitiateResult{}, ErrMissingRepository
	}
	if idempotencyKey == "" {
		return InitiateResult{}, ErrMissingIdempotencyKey
	}

	file, err := ValidateInitiate(fileName, size)
	if err != nil {
		return InitiateResult{}, err
	}

	now := s.now().UTC()
	existing, err := s.sessions.GetByIdempotencyKey(ctx, ownerID, idempotencyKey)
	if err == nil {
		if existing.Status != SessionInitiated || existing.UploadMode != uploadMode || !now.Before(existing.ExpiresAt) {
			return InitiateResult{}, ErrIdempotencyConflict
		}
		uploadURL, signingErr := s.generateUploadURL(uploadMode, existing.BlobName, existing.ExpiresAt)
		if signingErr != nil {
			return InitiateResult{}, fmt.Errorf("%w: %w", ErrGenerateUploadURL, signingErr)
		}
		return newInitiateResult(existing.ID, existing.BlobName, uploadURL, existing.ExpiresAt, uploadMode), nil
	}
	if !errors.Is(err, ErrSessionNotFound) {
		return InitiateResult{}, fmt.Errorf("%w: %w", ErrGetIdempotentSession, err)
	}

	id, err := s.newID()
	if err != nil {
		return InitiateResult{}, err
	}

	createdAt := now
	expiresAt := now.Add(URLLifetime)
	blobName := id + file.Extension
	if err := s.sessions.Create(ctx, UploadSession{
		ID:             id,
		OwnerID:        ownerID,
		DisplayName:    file.Name,
		BlobName:       blobName,
		DeclaredSize:   file.Size,
		Extension:      file.Extension,
		ContentType:    file.ContentType,
		UploadMode:     uploadMode,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         SessionInitiated,
		Version:        1,
	}); err != nil {
		return InitiateResult{}, fmt.Errorf("%w: %w", ErrCreateSession, err)
	}

	uploadURL, err := s.generateUploadURL(uploadMode, blobName, expiresAt)
	if err != nil {
		return InitiateResult{}, fmt.Errorf("%w: %w", ErrGenerateUploadURL, err)
	}

	return newInitiateResult(id, blobName, uploadURL, expiresAt, uploadMode), nil
}

func (s *Service) generateUploadURL(uploadMode UploadMode, blobName string, expiresAt time.Time) (string, error) {
	if uploadMode == UploadModeChunked {
		return s.signer.GenerateBlockUploadURL(blobName, expiresAt)
	}

	return s.signer.GenerateUploadURL(blobName, expiresAt)
}

func newInitiateResult(id string, blobName string, uploadURL string, expiresAt time.Time, uploadMode UploadMode) InitiateResult {
	result := InitiateResult{ID: id, BlobName: blobName, UploadURL: uploadURL, ExpiresAt: expiresAt, UploadMode: uploadMode}
	if uploadMode == UploadModeChunked {
		result.BlockSize = ChunkBlockSize
	}

	return result
}

func (s *Service) Get(ctx context.Context, ownerID string, id string) (SessionResult, error) {
	if ownerID == "" {
		return SessionResult{}, ErrMissingOwner
	}
	if s.sessions == nil {
		return SessionResult{}, ErrMissingRepository
	}

	session, err := s.sessions.GetByID(ctx, id)
	if err != nil {
		return SessionResult{}, err
	}
	if session.OwnerID != ownerID {
		return SessionResult{}, ErrSessionForbidden
	}

	return SessionResult{
		ID:        session.ID,
		FileName:  session.DisplayName,
		Size:      session.DeclaredSize,
		Status:    session.Status,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (s *Service) Complete(ctx context.Context, ownerID string, id string) (SessionResult, error) {
	if ownerID == "" {
		return SessionResult{}, ErrMissingOwner
	}
	if s.sessions == nil {
		return SessionResult{}, ErrMissingRepository
	}
	if s.inspector == nil {
		return SessionResult{}, ErrMissingBlobInspector
	}

	session, err := s.sessions.GetByID(ctx, id)
	if err != nil {
		return SessionResult{}, err
	}
	if session.OwnerID != ownerID {
		return SessionResult{}, ErrSessionForbidden
	}
	if session.Status == SessionUploaded {
		return sessionResult(session), nil
	}
	if !s.now().UTC().Before(session.ExpiresAt) {
		return SessionResult{}, ErrSessionExpired
	}
	if !session.Status.CanTransitionTo(SessionUploaded) {
		return SessionResult{}, ErrInvalidCompletionState
	}

	properties, err := s.inspector.GetBlobProperties(ctx, session.BlobName)
	if err != nil {
		if errors.Is(err, ErrBlobNotFound) {
			return SessionResult{}, ErrBlobNotFound
		}
		return SessionResult{}, err
	}
	if properties.Size != session.DeclaredSize {
		if _, err := s.sessions.UpdateStatus(ctx, session.ID, session.Version, SessionRejected); err != nil {
			return SessionResult{}, err
		}
		return SessionResult{}, ErrBlobSizeMismatch
	}

	updated, err := s.sessions.UpdateStatus(ctx, session.ID, session.Version, SessionUploaded)
	if err == nil {
		return sessionResult(updated), nil
	}
	if !errors.Is(err, ErrSessionVersionConflict) {
		return SessionResult{}, err
	}

	updated, err = s.sessions.GetByID(ctx, session.ID)
	if err != nil {
		return SessionResult{}, err
	}
	if updated.OwnerID == ownerID && updated.Status == SessionUploaded {
		return sessionResult(updated), nil
	}

	return SessionResult{}, ErrSessionVersionConflict
}

func sessionResult(session UploadSession) SessionResult {
	return SessionResult{
		ID:        session.ID,
		FileName:  session.DisplayName,
		Size:      session.DeclaredSize,
		Status:    session.Status,
		ExpiresAt: session.ExpiresAt,
	}
}

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("%w: %w", ErrGenerateID, err)
	}

	return hex.EncodeToString(bytes[:]), nil
}
