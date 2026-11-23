package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nshathish/uploadhub/backend/internal/upload"
)

type fakeUploadInitiator struct {
	result upload.InitiateResult
	err    error
	calls  int
}

func (f *fakeUploadInitiator) Initiate(_ context.Context, _ string, _ string, _ string, _ int64) (upload.InitiateResult, error) {
	f.calls++
	return f.result, f.err
}

func (f *fakeUploadInitiator) InitiateChunked(_ context.Context, _ string, _ string, _ string, _ int64) (upload.InitiateResult, error) {
	f.calls++
	return f.result, f.err
}

func TestInitiateUploadHandler(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantCalls   int
	}{
		{name: "rejects non JSON request", contentType: "text/plain", body: "not json", wantStatus: http.StatusUnsupportedMediaType},
		{name: "rejects unknown JSON field", contentType: "application/json", body: `{"fileName":"document.pdf","size":1,"unexpected":true}`, wantStatus: http.StatusBadRequest},
		{name: "accepts valid request with content type parameters", contentType: "application/json; charset=utf-8", body: `{"fileName":"document.pdf","size":1}`, wantStatus: http.StatusCreated, wantCalls: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUploadInitiator{result: upload.InitiateResult{ID: "opaque-id", BlobName: "opaque-id.pdf", UploadURL: "https://storage.invalid/upload", ExpiresAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}}
			request := httptest.NewRequest(http.MethodPost, "/api/uploads/initiate", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			request.Header.Set("Idempotency-Key", "test-key")
			request = request.WithContext(context.WithValue(request.Context(), ownerContextKey{}, "demo-user"))
			recorder := httptest.NewRecorder()

			NewInitiateUploadHandler(fake).ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %q", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if fake.calls != test.wantCalls {
				t.Fatalf("initiate calls = %d, want %d", fake.calls, test.wantCalls)
			}
		})
	}
}

func TestInitiateChunkedUploadHandler(t *testing.T) {
	fake := &fakeUploadInitiator{result: upload.InitiateResult{ID: "opaque-id", UploadMode: upload.UploadModeChunked, BlockSize: upload.ChunkBlockSize}}
	request := httptest.NewRequest(http.MethodPost, "/api/uploads/initiate-chunked", strings.NewReader(`{"fileName":"document.pdf","size":1}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "test-key")
	request = request.WithContext(context.WithValue(request.Context(), ownerContextKey{}, "demo-user"))
	recorder := httptest.NewRecorder()

	NewInitiateChunkedUploadHandler(fake).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}
