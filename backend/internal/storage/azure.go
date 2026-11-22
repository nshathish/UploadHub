package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

var (
	ErrMissingConnectionString = errors.New("missing Azure Storage connection string")
	ErrMissingContainerName    = errors.New("missing Azure Storage container name")
	ErrCreateBlobClient        = errors.New("create Azure Blob client")
	ErrGenerateSASURL          = errors.New("generate Azure Blob SAS URL")
)

type AzureBlobSigner struct {
	connectionString string
	containerName    string
}

func NewAzureBlobSigner(connectionString string, containerName string) (*AzureBlobSigner, error) {
	if connectionString == "" {
		return nil, ErrMissingConnectionString
	}
	if containerName == "" {
		return nil, ErrMissingContainerName
	}

	return &AzureBlobSigner{
		connectionString: connectionString,
		containerName:    containerName,
	}, nil
}

func (s *AzureBlobSigner) GenerateUploadURL(blobName string, expiresAt time.Time) (string, error) {
	client, err := blob.NewClientFromConnectionString(s.connectionString, s.containerName, blobName, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCreateBlobClient, err)
	}

	startTime := time.Now().UTC().Add(-5 * time.Minute)
	uploadURL, err := client.GetSASURL(sas.BlobPermissions{Create: true}, expiresAt, &blob.GetSASURLOptions{StartTime: &startTime})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGenerateSASURL, err)
	}

	return uploadURL, nil
}
