package upload

import (
	"errors"
	"testing"
)

func TestValidateInitiate(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		size     int64
		wantErr  error
		wantType string
	}{
		{name: "accepts PDF", fileName: "passport.PDF", size: 1024, wantType: "application/pdf"},
		{name: "accepts JPEG", fileName: "address.jpg", size: 1024, wantType: "image/jpeg"},
		{name: "rejects path", fileName: "../passport.pdf", size: 1024, wantErr: ErrInvalidFileName},
		{name: "rejects empty name", fileName: "", size: 1024, wantErr: ErrInvalidFileName},
		{name: "rejects empty file", fileName: "passport.pdf", size: 0, wantErr: ErrInvalidFileSize},
		{name: "rejects oversized file", fileName: "passport.pdf", size: MaxFileSize + 1, wantErr: ErrInvalidFileSize},
		{name: "rejects unsupported file", fileName: "passport.exe", size: 1024, wantErr: ErrUnsupportedType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := ValidateInitiate(test.fileName, test.size)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ValidateInitiate() error = %v, want %v", err, test.wantErr)
			}
			if err == nil && file.ContentType != test.wantType {
				t.Errorf("ContentType = %q, want %q", file.ContentType, test.wantType)
			}
		})
	}
}
