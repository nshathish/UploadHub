package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrLoadEnvironment               = errors.New("load environment")
	ErrMissingAzureStorageConnection = errors.New("missing Azure Storage connection string")
)

type Config struct {
	Port                  string
	AzureConnectionString string
	AzureContainerName    string
	UploadDatabasePath    string
	DevelopmentAuthToken  string
	DevelopmentOwnerID    string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("%w: %w", ErrLoadEnvironment, err)
	}

	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connectionString == "" {
		return Config{}, ErrMissingAzureStorageConnection
	}

	return Config{
		Port:                  valueOrDefault("APP_PORT", "8080"),
		AzureConnectionString: connectionString,
		AzureContainerName:    valueOrDefault("AZURE_STORAGE_CONTAINER", "kyc-documents"),
		UploadDatabasePath:    valueOrDefault("UPLOAD_DATABASE_PATH", "uploadhub.db"),
		DevelopmentAuthToken:  valueOrDefault("UPLOADHUB_DEV_AUTH_TOKEN", "local-development-token"),
		DevelopmentOwnerID:    valueOrDefault("UPLOADHUB_DEV_OWNER_ID", "local-demo-user"),
	}, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
