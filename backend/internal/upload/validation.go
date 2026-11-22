package upload

import (
	"errors"
	"path"
	"strings"
)

const MaxFileSize int64 = 25 << 20

var (
	ErrInvalidFileName = errors.New("invalid file name")
	ErrInvalidFileSize = errors.New("invalid file size")
	ErrUnsupportedType = errors.New("unsupported file type")
)

type FileDetails struct {
	Name        string
	Size        int64
	Extension   string
	ContentType string
}

func ValidateInitiate(fileName string, size int64) (FileDetails, error) {
	if fileName == "" || fileName == "." || strings.ContainsAny(fileName, `/\\`) {
		return FileDetails{}, ErrInvalidFileName
	}

	if size <= 0 || size > MaxFileSize {
		return FileDetails{}, ErrInvalidFileSize
	}

	extension := strings.ToLower(path.Ext(fileName))
	contentType, ok := allowedContentTypes[extension]
	if !ok {
		return FileDetails{}, ErrUnsupportedType
	}

	return FileDetails{
		Name:        fileName,
		Size:        size,
		Extension:   extension,
		ContentType: contentType,
	}, nil
}

var allowedContentTypes = map[string]string{
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
}
