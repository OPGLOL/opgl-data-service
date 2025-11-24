package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nathannewyen/opgl-data/internal/services"
)

// Handler manages HTTP request handlers for the data service
type Handler struct {
	riotService *services.RiotService
}

// NewHandler creates a new Handler instance
func NewHandler(riotService *services.RiotService) *Handler {
	return &Handler{
		riotService: riotService,
	}
}

// HealthCheck handles health check requests
func (handler *Handler) HealthCheck(writer http.ResponseWriter, request *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "opgl-data",
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

// GetSummoner handles summoner lookup requests
func (handler *Handler) GetSummoner(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	region := vars["region"]
	summonerName := vars["summonerName"]

	summoner, err := handler.riotService.GetSummonerByName(region, summonerName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(summoner)
}

// GetMatches handles match history requests
func (handler *Handler) GetMatches(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	region := vars["region"]
	puuid := vars["puuid"]

	// Get count from query parameter (default: 20)
	countStr := request.URL.Query().Get("count")
	count := 20
	if countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 {
			count = parsedCount
		}
	}

	matches, err := handler.riotService.GetMatchHistory(region, puuid, count)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(matches)
}
