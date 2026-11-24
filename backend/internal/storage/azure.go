package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/nshathish/uploadhub/backend/internal/upload"
)

var (
	ErrMissingConnectionString = errors.New("missing Azure Storage connection string")
	ErrMissingContainerName    = errors.New("missing Azure Storage container name")
	ErrCreateBlobClient        = errors.New("create Azure Blob client")
	ErrGenerateSASURL          = errors.New("generate Azure Blob SAS URL")
	ErrGetBlobProperties       = errors.New("get Azure Blob properties")
	ErrInvalidBlobProperties   = errors.New("invalid Azure Blob properties")
	ErrGetBlockList            = errors.New("get Azure Blob block list")
)

type AzureBlobStorage struct {
	connectionString string
	containerName    string
}

func (s *AzureBlobStorage) GetUncommittedBlockIDs(ctx context.Context, blobName string) ([]string, error) {
	client, err := blockblob.NewClientFromConnectionString(s.connectionString, s.containerName, blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBlobClient, err)
	}
	blockList, err := client.GetBlockList(ctx, blockblob.BlockListTypeUncommitted, nil)
	if err != nil {
		var responseError *azcore.ResponseError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return []string{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrGetBlockList, err)
	}
	ids := make([]string, 0, len(blockList.UncommittedBlocks))
	for _, block := range blockList.UncommittedBlocks {
		if block.Name != nil {
			ids = append(ids, *block.Name)
		}
	}
	return ids, nil
}

func NewAzureBlobStorage(connectionString string, containerName string) (*AzureBlobStorage, error) {
	if connectionString == "" {
		return nil, ErrMissingConnectionString
	}
	if containerName == "" {
		return nil, ErrMissingContainerName
	}

	return &AzureBlobStorage{
		connectionString: connectionString,
		containerName:    containerName,
	}, nil
}

func (s *AzureBlobStorage) GenerateUploadURL(blobName string, expiresAt time.Time) (string, error) {
	return s.generateSASURL(blobName, expiresAt, sas.BlobPermissions{Create: true})
}

func (s *AzureBlobStorage) GenerateBlockUploadURL(blobName string, expiresAt time.Time) (string, error) {
	return s.generateSASURL(blobName, expiresAt, sas.BlobPermissions{Write: true})
}

func (s *AzureBlobStorage) generateSASURL(blobName string, expiresAt time.Time, permissions sas.BlobPermissions) (string, error) {
	client, err := blob.NewClientFromConnectionString(s.connectionString, s.containerName, blobName, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateBlobClient, err)
	}

	startTime := time.Now().UTC().Add(-5 * time.Minute)
	uploadURL, err := client.GetSASURL(permissions, expiresAt, &blob.GetSASURLOptions{StartTime: &startTime})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGenerateSASURL, err)
	}

	return uploadURL, nil
}

func (s *AzureBlobStorage) GetBlobProperties(ctx context.Context, blobName string) (upload.BlobProperties, error) {
	client, err := blob.NewClientFromConnectionString(s.connectionString, s.containerName, blobName, nil)
	if err != nil {
		return upload.BlobProperties{}, fmt.Errorf("%w: %w", ErrCreateBlobClient, err)
	}

	properties, err := client.GetProperties(ctx, nil)
	if err != nil {
		var responseError *azcore.ResponseError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return upload.BlobProperties{}, fmt.Errorf("%w: %w", upload.ErrBlobNotFound, err)
		}
		return upload.BlobProperties{}, fmt.Errorf("%w: %w", ErrGetBlobProperties, err)
	}
	if properties.ContentLength == nil {
		return upload.BlobProperties{}, ErrInvalidBlobProperties
	}

	result := upload.BlobProperties{Size: *properties.ContentLength}
	if properties.ContentType != nil {
		result.ContentType = *properties.ContentType
	}
	if properties.ETag != nil {
		result.ETag = string(*properties.ETag)
	}

	return result, nil
}
