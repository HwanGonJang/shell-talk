package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	PostgresURL string
	MongoURL    string
}

// Load loads configuration from environment variables.
func Load() *Config {
	// Load .env file if it exists (useful for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		postgresURL = "postgres://user:password@localhost:5432/shelltalk?sslmode=disable"
	}

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://user:password@localhost:27017"
	}

	return &Config{
		PostgresURL: postgresURL,
		MongoURL:    mongoURL,
	}
}
