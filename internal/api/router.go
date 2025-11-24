package api

import (
	"github.com/gorilla/mux"
)

// SetupRouter configures all routes for the data service
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// Data endpoints
	router.HandleFunc("/api/v1/summoner/{region}/{summonerName}", handler.GetSummoner).Methods("GET")
	router.HandleFunc("/api/v1/matches/{region}/{puuid}", handler.GetMatches).Methods("GET")

	return router
}
