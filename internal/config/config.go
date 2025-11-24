package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the OPGL application
type Config struct {
	// Riot Games API key for accessing League of Legends data
	RiotAPIKey string
	// Server port number for the HTTP server
	ServerPort string
	// Database connection string (for future persistence)
	DatabaseURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	// Error is ignored because .env file is optional (production might use actual env vars)
	_ = godotenv.Load()

	riotAPIKey := os.Getenv("RIOT_API_KEY")
	if riotAPIKey == "" {
		log.Println("Warning: RIOT_API_KEY not set. API calls will fail.")
	}

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8081"
	}

	databaseURL := os.Getenv("DATABASE_URL")

	return &Config{
		RiotAPIKey:  riotAPIKey,
		ServerPort:  serverPort,
		DatabaseURL: databaseURL,
	}
}
