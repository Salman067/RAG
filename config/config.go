package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIAPIKey string
	QdrantURL    string
	DatabaseURL  string
	// Add other config fields
}

func LoadConfig() *Config {
	godotenv.Load()
	return &Config{
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		QdrantURL:    "localhost:6334", // Default Qdrant URL
		DatabaseURL:  os.Getenv("DATABASE_URL"),
	}
}
