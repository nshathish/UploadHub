package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nshathish/uploadhub/backend/internal/config"
	"github.com/nshathish/uploadhub/backend/internal/handlers"
	"github.com/nshathish/uploadhub/backend/internal/storage"
	"github.com/nshathish/uploadhub/backend/internal/upload"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	blobStorage, err := storage.NewAzureBlobStorage(cfg.AzureConnectionString, cfg.AzureContainerName)
	if err != nil {
		log.Fatal(err)
	}

	sessions, err := upload.NewSQLiteSessionRepository(cfg.UploadDatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer sessions.Close()
	uploadService := upload.NewService(blobStorage, blobStorage, sessions)

	mux := http.NewServeMux()
	mux.Handle(
		"POST /api/uploads/initiate",
		handlers.DevelopmentAuth(
			cfg.DevelopmentAuthToken,
			cfg.DevelopmentOwnerID,
			handlers.NewInitiateUploadHandler(uploadService),
		),
	)
	mux.Handle(
		"POST /api/uploads/initiate-chunked",
		handlers.DevelopmentAuth(
			cfg.DevelopmentAuthToken,
			cfg.DevelopmentOwnerID,
			handlers.NewInitiateChunkedUploadHandler(uploadService),
		),
	)
	mux.Handle(
		"GET /api/uploads/{id}",
		handlers.DevelopmentAuth(
			cfg.DevelopmentAuthToken,
			cfg.DevelopmentOwnerID,
			handlers.NewGetUploadHandler(uploadService),
		),
	)
	mux.Handle(
		"POST /api/uploads/{id}/complete",
		handlers.DevelopmentAuth(
			cfg.DevelopmentAuthToken,
			cfg.DevelopmentOwnerID,
			handlers.NewCompleteUploadHandler(uploadService),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("UploadHub API listening on http://localhost:%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
