package config

import (
	"fmt"
	"os"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	Storage     StorageConfig
	StaticDir   string
}

// StorageConfig holds S3-compatible object storage settings.
type StorageConfig struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
}

// Load reads configuration from the environment.
func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	storage := StorageConfig{
		Endpoint:        os.Getenv("STORAGE_ENDPOINT"),
		Region:          getEnvDefault("STORAGE_REGION", "us-east-1"),
		Bucket:          os.Getenv("STORAGE_BUCKET"),
		AccessKeyID:     os.Getenv("STORAGE_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("STORAGE_SECRET_ACCESS_KEY"),
		UsePathStyle:    os.Getenv("STORAGE_USE_PATH_STYLE") == "true",
	}

	if storage.Endpoint == "" {
		return Config{}, fmt.Errorf("STORAGE_ENDPOINT is required")
	}
	if storage.Bucket == "" {
		return Config{}, fmt.Errorf("STORAGE_BUCKET is required")
	}
	if storage.AccessKeyID == "" {
		return Config{}, fmt.Errorf("STORAGE_ACCESS_KEY_ID is required")
	}
	if storage.SecretAccessKey == "" {
		return Config{}, fmt.Errorf("STORAGE_SECRET_ACCESS_KEY is required")
	}

	staticDir := getEnvDefault("STATIC_DIR", "web/dist")

	return Config{
		Port:        port,
		DatabaseURL: dbURL,
		Storage:     storage,
		StaticDir:   staticDir,
	}, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
