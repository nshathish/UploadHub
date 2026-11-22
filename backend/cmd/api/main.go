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

	signer, err := storage.NewAzureBlobSigner(cfg.AzureConnectionString, cfg.AzureContainerName)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle(
		"POST /api/uploads/initiate",
		handlers.DevelopmentAuth(
			cfg.DevelopmentAuthToken,
			cfg.DevelopmentOwnerID,
			handlers.NewInitiateUploadHandler(upload.NewService(signer)),
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
