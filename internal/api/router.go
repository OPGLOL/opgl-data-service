package api

import (
	"github.com/gorilla/mux"
)

// SetupRouter configures all routes for the data service
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", handler.HealthCheck).Methods("POST")

	// Data endpoints
	router.HandleFunc("/api/v1/summoner", handler.GetSummonerByRiotID).Methods("POST")
	router.HandleFunc("/api/v1/matches", handler.GetMatchesByRiotID).Methods("POST")

	return router
}
