package upload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const URLLifetime = 10 * time.Minute

type URLSigner interface {
	GenerateUploadURL(ctx context.Context, blobName string, expiresAt time.Time) (string, error)
}

type InitiateResult struct {
	ID        string    `json:"id"`
	BlobName  string    `json:"blobName"`
	UploadURL string    `json:"uploadUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Service struct {
	signer URLSigner
	now    func() time.Time
	newID  func() (string, error)
}

func NewService(signer URLSigner) *Service {
	return &Service{
		signer: signer,
		now:    time.Now,
		newID:  newID,
	}
}

func (s *Service) Initiate(ctx context.Context, fileName string, size int64) (InitiateResult, error) {
	file, err := ValidateInitiate(fileName, size)
	if err != nil {
		return InitiateResult{}, err
	}

	id, err := s.newID()
	if err != nil {
		return InitiateResult{}, err
	}

	expiresAt := s.now().UTC().Add(URLLifetime)
	blobName := id + file.Extension
	uploadURL, err := s.signer.GenerateUploadURL(ctx, blobName, expiresAt)
	if err != nil {
		return InitiateResult{}, fmt.Errorf("generate upload URL: %w", err)
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
		return "", fmt.Errorf("generate upload ID: %w", err)
	}

	return hex.EncodeToString(bytes[:]), nil
}
