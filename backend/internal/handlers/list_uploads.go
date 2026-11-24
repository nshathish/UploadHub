package handlers

import (
	"context"
	"net/http"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type UploadSessionLister interface {
	List(ctx context.Context, ownerID string) ([]upload.SessionResult, error)
}

func NewListUploadsHandler(service UploadSessionLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, ok := OwnerID(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		results, err := service.List(r.Context(), ownerID)
		if err != nil {
			http.Error(w, "unable to retrieve uploads", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, results)
	}
}
