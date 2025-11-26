package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/OPGLOL/opgl-data-service/internal/api"
	"github.com/OPGLOL/opgl-data-service/internal/config"
	"github.com/OPGLOL/opgl-data-service/internal/middleware"
	"github.com/OPGLOL/opgl-data-service/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize zerolog with colorized console output for development
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Caller().Logger()

	// Set global log level (can be configured via environment variable)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Starting OPGL Data Service")

	// Load configuration
	configuration := config.LoadConfig()

	log.Info().
		Str("port", configuration.ServerPort).
		Bool("riot_api_key_set", configuration.RiotAPIKey != "").
		Msg("Configuration loaded")

	// Initialize Riot service
	riotService := services.NewRiotService(configuration.RiotAPIKey)

	// Initialize HTTP handler
	handler := api.NewHandler(riotService)

	// Set up router
	router := api.SetupRouter(handler)

	// Wrap router with logging middleware
	loggedRouter := middleware.LoggingMiddleware(router)

	// Start server
	serverAddress := fmt.Sprintf(":%s", configuration.ServerPort)
	log.Info().
		Str("address", serverAddress).
		Str("port", configuration.ServerPort).
		Msg("OPGL Data Service listening")

	if err := http.ListenAndServe(serverAddress, loggedRouter); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
