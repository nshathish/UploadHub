package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type UploadInitiator interface {
	Initiate(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64) (upload.InitiateResult, error)
}

type ChunkedUploadInitiator interface {
	InitiateChunked(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64) (upload.InitiateResult, error)
}

type initiateUploadFunc func(ctx context.Context, ownerID string, idempotencyKey string, fileName string, size int64) (upload.InitiateResult, error)

type InitiateUploadRequest struct {
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
}

func NewInitiateUploadHandler(service UploadInitiator) http.HandlerFunc {
	return newInitiateUploadHandler(service.Initiate)
}

func NewInitiateChunkedUploadHandler(service ChunkedUploadInitiator) http.HandlerFunc {
	return newInitiateUploadHandler(service.InitiateChunked)
}

func newInitiateUploadHandler(initiate initiateUploadFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var request InitiateUploadRequest
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "invalid JSON request", http.StatusBadRequest)
			return
		}

		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			http.Error(w, "unexpected request content", http.StatusBadRequest)
			return
		}

		ownerID, ok := OwnerID(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			http.Error(w, "idempotency key is required", http.StatusBadRequest)
			return
		}

		result, err := initiate(r.Context(), ownerID, idempotencyKey, request.FileName, request.Size)
		if err != nil {
			switch {
			case errors.Is(err, upload.ErrInvalidFileName), errors.Is(err, upload.ErrInvalidFileSize):
				http.Error(w, "invalid upload request", http.StatusBadRequest)
			case errors.Is(err, upload.ErrFileTooLarge):
				http.Error(w, "file exceeds chunked upload capacity", http.StatusRequestEntityTooLarge)
			case errors.Is(err, upload.ErrUnsupportedType):
				http.Error(w, "unsupported file type", http.StatusUnsupportedMediaType)
			case errors.Is(err, upload.ErrIdempotencyConflict):
				http.Error(w, "idempotency key cannot be reused", http.StatusConflict)
			default:
				http.Error(w, "unable to initiate upload", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusCreated, result)
	}
}
