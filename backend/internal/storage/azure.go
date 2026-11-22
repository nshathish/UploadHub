package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

type AzureBlobSigner struct {
	connectionString string
	containerName    string
}

func NewAzureBlobSigner(connectionString string, containerName string) (*AzureBlobSigner, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("Azure Storage connection string is required")
	}
	if containerName == "" {
		return nil, fmt.Errorf("Azure Storage container name is required")
	}

	return &AzureBlobSigner{
		connectionString: connectionString,
		containerName:    containerName,
	}, nil
}

func (s *AzureBlobSigner) GenerateUploadURL(_ context.Context, blobName string, expiresAt time.Time) (string, error) {
	client, err := blob.NewClientFromConnectionString(s.connectionString, s.containerName, blobName, nil)
	if err != nil {
		return "", fmt.Errorf("create Azure Blob client: %w", err)
	}

	startTime := time.Now().UTC().Add(-5 * time.Minute)
	uploadURL, err := client.GetSASURL(sas.BlobPermissions{Create: true}, expiresAt, &blob.GetSASURLOptions{StartTime: &startTime})
	if err != nil {
		return "", fmt.Errorf("generate Azure Blob SAS URL: %w", err)
	}

	return uploadURL, nil
}
