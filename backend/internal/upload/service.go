package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const URLLifetime = 10 * time.Minute

var (
	ErrGenerateUploadURL     = errors.New("generate upload URL")
	ErrGenerateID            = errors.New("generate upload ID")
	ErrMissingOwner          = errors.New("missing upload owner")
	ErrMissingRepository     = errors.New("missing upload session repository")
	ErrMissingIdempotencyKey = errors.New("missing idempotency key")
	ErrCreateSession         = errors.New("create upload session")
	ErrGetIdempotentSession  = errors.New("get idempotent upload session")
	ErrSessionForbidden      = errors.New("upload session access denied")
)

type URLSigner interface {
	GenerateUploadURL(blobName string, expiresAt time.Time) (string, error)
}

type InitiateResult struct {
	ID        string    `json:"id"`
	BlobName  string    `json:"blobName"`
	UploadURL string    `json:"uploadUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type SessionResult struct {
	ID        string        `json:"id"`
	FileName  string        `json:"fileName"`
	Size      int64         `json:"size"`
	Status    SessionStatus `json:"status"`
	ExpiresAt time.Time     `json:"expiresAt"`
}

type Service struct {
	signer   URLSigner
	sessions SessionRepository
	now      func() time.Time
	newID    func() (string, error)
}

func NewService(signer URLSigner, sessions SessionRepository) *Service {
	return &Service{
		signer:   signer,
		sessions: sessions,
		now:      time.Now,
		newID:    newID,
	}
}

func (s *Service) Initiate(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64) (InitiateResult, error) {
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
		if existing.Status != SessionInitiated || !now.Before(existing.ExpiresAt) {
			return InitiateResult{}, ErrIdempotencyConflict
		}
		uploadURL, signingErr := s.signer.GenerateUploadURL(existing.BlobName, existing.ExpiresAt)
		if signingErr != nil {
			return InitiateResult{}, fmt.Errorf("%w: %w", ErrGenerateUploadURL, signingErr)
		}
		return InitiateResult{ID: existing.ID, BlobName: existing.BlobName, UploadURL: uploadURL, ExpiresAt: existing.ExpiresAt}, nil
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
		IdempotencyKey: idempotencyKey,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         SessionInitiated,
		Version:        1,
	}); err != nil {
		return InitiateResult{}, fmt.Errorf("%w: %w", ErrCreateSession, err)
	}

	uploadURL, err := s.signer.GenerateUploadURL(blobName, expiresAt)
	if err != nil {
		return InitiateResult{}, fmt.Errorf("%w: %w", ErrGenerateUploadURL, err)
	}

	return InitiateResult{
		ID:        id,
		BlobName:  blobName,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
	}, nil
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

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("%w: %w", ErrGenerateID, err)
	}

	return hex.EncodeToString(bytes[:]), nil
}
