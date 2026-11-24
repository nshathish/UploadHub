package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type ChunkedUploadResumer interface {
	ResumeChunked(ctx context.Context, ownerID string, id string) (upload.ChunkResumeResult, error)
}

func NewResumeChunkedUploadHandler(service ChunkedUploadResumer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, ok := OwnerID(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		result, err := service.ResumeChunked(r.Context(), ownerID, r.PathValue("id"))
		if err != nil {
			switch {
			case errors.Is(err, upload.ErrSessionNotFound), errors.Is(err, upload.ErrSessionForbidden):
				http.Error(w, "upload not found", http.StatusNotFound)
			case errors.Is(err, upload.ErrInvalidResumeState):
				http.Error(w, "upload cannot be resumed", http.StatusConflict)
			default:
				http.Error(w, "unable to resume upload", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, result)
	}
}
