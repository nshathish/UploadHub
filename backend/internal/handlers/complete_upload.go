package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type UploadCompleter interface {
	Complete(ctx context.Context, ownerID string, id string) (upload.SessionResult, error)
}

func NewCompleteUploadHandler(service UploadCompleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, ok := OwnerID(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		result, err := service.Complete(r.Context(), ownerID, r.PathValue("id"))
		if err != nil {
			switch {
			case errors.Is(err, upload.ErrSessionNotFound), errors.Is(err, upload.ErrSessionForbidden):
				http.Error(w, "upload not found", http.StatusNotFound)
			case errors.Is(err, upload.ErrBlobNotFound):
				http.Error(w, "uploaded blob not found", http.StatusConflict)
			case errors.Is(err, upload.ErrBlobSizeMismatch):
				http.Error(w, "uploaded blob size does not match", http.StatusUnprocessableEntity)
			case errors.Is(err, upload.ErrSessionExpired), errors.Is(err, upload.ErrInvalidCompletionState), errors.Is(err, upload.ErrSessionVersionConflict):
				http.Error(w, "upload cannot be completed", http.StatusConflict)
			default:
				http.Error(w, "unable to complete upload", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, result)
	}
}
