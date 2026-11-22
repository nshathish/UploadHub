package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                  string
	AzureConnectionString string
	AzureContainerName    string
}

func Load() (Config, error) {
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connectionString == "" {
		return Config{}, fmt.Errorf("AZURE_STORAGE_CONNECTION_STRING is required")
	}

	return Config{
		Port:                  valueOrDefault("APP_PORT", "8080"),
		AzureConnectionString: connectionString,
		AzureContainerName:    valueOrDefault("AZURE_STORAGE_CONTAINER", "kyc-documents"),
	}, nil
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
