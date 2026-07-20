package main

import (
	"errors"
	"log"
	"os"

	"github.com/aadielpr/unnamed/internal/config"
	"github.com/aadielpr/unnamed/internal/db"
	"github.com/aadielpr/unnamed/internal/server"
	"github.com/aadielpr/unnamed/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	if err := loadEnv(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer database.Close()

	_, err = storage.NewS3Store(storage.Config{
		Endpoint:        cfg.Storage.Endpoint,
		Region:          cfg.Storage.Region,
		Bucket:          cfg.Storage.Bucket,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		UsePathStyle:    cfg.Storage.UsePathStyle,
	})
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	srv := server.New(cfg, database)
	log.Printf("server listening on :%s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
