package upload

import (
	"errors"
	"path"
	"strings"
)

const (
	MaxSingleFileSize  int64 = 25 << 20
	ChunkBlockSize     int64 = 4 << 20
	MaxChunkBlockSize  int64 = 4000 << 20
	MaxChunkBlockCount int64 = 50000
)

var (
	ErrInvalidFileName = errors.New("invalid file name")
	ErrInvalidFileSize = errors.New("invalid file size")
	ErrFileTooLarge    = errors.New("file exceeds Azure block blob capacity")
	ErrUnsupportedType = errors.New("unsupported file type")
)

type FileDetails struct {
	Name        string
	Size        int64
	Extension   string
	ContentType string
}

func ValidateInitiate(fileName string, size int64) (FileDetails, error) {
	file, err := validateFile(fileName, size)
	if err != nil {
		return FileDetails{}, err
	}
	if size > MaxSingleFileSize {
		return FileDetails{}, ErrInvalidFileSize
	}

	return file, nil
}

func ValidateChunkedInitiate(fileName string, size int64) (FileDetails, int64, error) {
	file, err := validateFile(fileName, size)
	if err != nil {
		return FileDetails{}, 0, err
	}

	blockSize := chunkBlockSizeFor(size)
	if blockSize > MaxChunkBlockSize {
		return FileDetails{}, 0, ErrFileTooLarge
	}

	return file, blockSize, nil
}

func validateFile(fileName string, size int64) (FileDetails, error) {
	if fileName == "" || fileName == "." || strings.ContainsAny(fileName, `/\\`) {
		return FileDetails{}, ErrInvalidFileName
	}

	if size <= 0 {
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

func chunkBlockSizeFor(size int64) int64 {
	minimumBlockSize := (size + MaxChunkBlockCount - 1) / MaxChunkBlockCount
	if minimumBlockSize <= ChunkBlockSize {
		return ChunkBlockSize
	}

	return ((minimumBlockSize + ChunkBlockSize - 1) / ChunkBlockSize) * ChunkBlockSize
}

var allowedContentTypes = map[string]string{
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
}
