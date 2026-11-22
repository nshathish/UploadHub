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
	ErrGenerateUploadURL = errors.New("generate upload URL")
	ErrGenerateID        = errors.New("generate upload ID")
	ErrMissingOwner      = errors.New("missing upload owner")
	ErrMissingRepository = errors.New("missing upload session repository")
	ErrCreateSession     = errors.New("create upload session")
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

func (s *Service) Initiate(ctx context.Context, ownerID string, fileName string, size int64) (InitiateResult, error) {
	if ownerID == "" {
		return InitiateResult{}, ErrMissingOwner
	}
	if s.sessions == nil {
		return InitiateResult{}, ErrMissingRepository
	}

	file, err := ValidateInitiate(fileName, size)
	if err != nil {
		return InitiateResult{}, err
	}

	id, err := s.newID()
	if err != nil {
		return InitiateResult{}, err
	}

	createdAt := s.now().UTC()
	expiresAt := createdAt.Add(URLLifetime)
	blobName := id + file.Extension
	if err := s.sessions.Create(ctx, UploadSession{
		ID:           id,
		OwnerID:      ownerID,
		DisplayName:  file.Name,
		BlobName:     blobName,
		DeclaredSize: file.Size,
		Extension:    file.Extension,
		ContentType:  file.ContentType,
		CreatedAt:    createdAt,
		ExpiresAt:    expiresAt,
		Status:       SessionInitiated,
		Version:      1,
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

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("%w: %w", ErrGenerateID, err)
	}

	return hex.EncodeToString(bytes[:]), nil
}
