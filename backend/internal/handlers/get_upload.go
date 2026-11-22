package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type UploadSessionReader interface {
	Get(ctx context.Context, ownerID string, id string) (upload.SessionResult, error)
}

func NewGetUploadHandler(service UploadSessionReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, ok := OwnerID(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		result, err := service.Get(r.Context(), ownerID, r.PathValue("id"))
		if err != nil {
			switch {
			case errors.Is(err, upload.ErrSessionNotFound), errors.Is(err, upload.ErrSessionForbidden):
				http.Error(w, "upload not found", http.StatusNotFound)
			default:
				http.Error(w, "unable to retrieve upload", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, result)
	}
}
