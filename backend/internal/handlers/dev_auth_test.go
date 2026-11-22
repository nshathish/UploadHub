package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevelopmentAuth(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		wantStatus    int
		wantOwnerID   string
	}{
		{name: "rejects missing token", wantStatus: http.StatusUnauthorized},
		{name: "rejects invalid token", authorization: "Bearer incorrect", wantStatus: http.StatusUnauthorized},
		{name: "accepts valid token", authorization: "Bearer development-token", wantStatus: http.StatusNoContent, wantOwnerID: "demo-user"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := DevelopmentAuth("development-token", "demo-user", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ownerID, ok := OwnerID(r.Context())
				if !ok {
					t.Fatal("owner ID was not added to request context")
				}
				if ownerID != test.wantOwnerID {
					t.Fatalf("owner ID = %q, want %q", ownerID, test.wantOwnerID)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodPost, "/api/uploads/initiate", nil)
			request.Header.Set("Authorization", test.authorization)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}
