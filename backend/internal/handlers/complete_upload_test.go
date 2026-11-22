package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type fakeUploadCompleter struct {
	result upload.SessionResult
	err    error
}

func (f *fakeUploadCompleter) Complete(_ context.Context, _ string, _ string) (upload.SessionResult, error) {
	return f.result, f.err
}

func TestCompleteUploadHandlerReturnsBlobMissingConflict(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("POST /api/uploads/{id}/complete", DevelopmentAuth("development-token", "demo-user", NewCompleteUploadHandler(&fakeUploadCompleter{err: upload.ErrBlobNotFound})))

	request := httptest.NewRequest(http.MethodPost, "/api/uploads/upload-id/complete", nil)
	request.Header.Set("Authorization", "Bearer development-token")
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}
