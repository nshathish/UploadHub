package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type fakeUploadSessionReader struct {
	result upload.SessionResult
	err    error
}

func (f *fakeUploadSessionReader) Get(_ context.Context, _ string, _ string) (upload.SessionResult, error) {
	return f.result, f.err
}

func TestGetUploadHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /api/uploads/{id}", DevelopmentAuth("development-token", "demo-user", NewGetUploadHandler(&fakeUploadSessionReader{
		result: upload.SessionResult{ID: "upload-id", Status: upload.SessionInitiated, ExpiresAt: time.Date(2026, time.November, 22, 12, 10, 0, 0, time.UTC)},
	})))

	request := httptest.NewRequest(http.MethodGet, "/api/uploads/upload-id", nil)
	request.Header.Set("Authorization", "Bearer development-token")
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
