package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nathannewyen/opgl-data/internal/api"
	"github.com/nathannewyen/opgl-data/internal/config"
	"github.com/nathannewyen/opgl-data/internal/services"
)

func main() {
	// Load configuration
	configuration := config.Load()

	// Initialize Riot service
	riotService := services.NewRiotService(configuration.RiotAPIKey)

	// Initialize HTTP handler
	handler := api.NewHandler(riotService)

	// Set up router
	router := api.SetupRouter(handler)

	// Start server
	serverAddress := fmt.Sprintf(":%s", configuration.Port)
	log.Printf("OPGL Data Service starting on port %s", configuration.Port)
	log.Fatal(http.ListenAndServe(serverAddress, router))
}
