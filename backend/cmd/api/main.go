package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/nshathish/uploadhub/backend/internal/auth"
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
	uploadService := upload.NewService(blobStorage, blobStorage, sessions).WithBlockInspector(blobStorage)
	authRepository, err := auth.NewSQLiteRepository(cfg.UploadDatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer authRepository.Close()
	authService := auth.NewService(authRepository)
	if cfg.SeedDevelopmentUsers {
		if err := auth.SeedDevelopmentUsers(context.Background(), authService); err != nil {
			log.Fatal(err)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("POST /api/auth/login", handlers.NewLoginHandler(authService))
	mux.Handle("POST /api/auth/logout", handlers.NewLogoutHandler(authService))
	mux.Handle(
		"POST /api/uploads/initiate",
		handlers.SessionAuth(authService, handlers.NewInitiateUploadHandler(uploadService)),
	)
	mux.Handle(
		"POST /api/uploads/initiate-chunked",
		handlers.SessionAuth(authService, handlers.NewInitiateChunkedUploadHandler(uploadService)),
	)
	mux.Handle(
		"GET /api/uploads",
		handlers.SessionAuth(authService, handlers.NewListUploadsHandler(uploadService)),
	)
	mux.Handle(
		"GET /api/uploads/{id}/chunks",
		handlers.SessionAuth(authService, handlers.NewResumeChunkedUploadHandler(uploadService)),
	)
	mux.Handle(
		"GET /api/uploads/{id}",
		handlers.SessionAuth(authService, handlers.NewGetUploadHandler(uploadService)),
	)
	mux.Handle(
		"POST /api/uploads/{id}/complete",
		handlers.SessionAuth(authService, handlers.NewCompleteUploadHandler(uploadService)),
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("UploadHub API listening on http://localhost:%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
