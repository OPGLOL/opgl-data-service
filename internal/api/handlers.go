package api

import (
	"encoding/json"
	"net/http"

	"github.com/OPGLOL/opgl-data-service/internal/services"
)

// Handler manages HTTP request handlers for the data service
type Handler struct {
	riotService services.RiotServiceInterface
}

// NewHandler creates a new Handler instance
func NewHandler(riotService services.RiotServiceInterface) *Handler {
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

// GetSummonerByRiotID handles summoner lookup by Riot ID with JSON body
func (handler *Handler) GetSummonerByRiotID(writer http.ResponseWriter, request *http.Request) {
	// Parse JSON request body
	var summonerRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	if err := json.NewDecoder(request.Body).Decode(&summonerRequest); err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if summonerRequest.Region == "" || summonerRequest.GameName == "" || summonerRequest.TagLine == "" {
		http.Error(writer, "region, gameName, and tagLine are required", http.StatusBadRequest)
		return
	}

	summoner, err := handler.riotService.GetSummonerByRiotID(summonerRequest.Region, summonerRequest.GameName, summonerRequest.TagLine)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(summoner)
}

// GetMatchesByRiotID handles match history requests using Riot ID or PUUID with JSON body
func (handler *Handler) GetMatchesByRiotID(writer http.ResponseWriter, request *http.Request) {
	// Parse JSON request body
	var matchRequest struct {
		Region   string `json:"region"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
		PUUID    string `json:"puuid"`
		Count    int    `json:"count"`
	}

	if err := json.NewDecoder(request.Body).Decode(&matchRequest); err != nil {
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields - either (gameName + tagLine) OR puuid must be provided
	if matchRequest.Region == "" {
		http.Error(writer, "region is required", http.StatusBadRequest)
		return
	}

	var puuid string

	// If PUUID is provided, use it directly (for internal gateway use)
	if matchRequest.PUUID != "" {
		puuid = matchRequest.PUUID
	} else if matchRequest.GameName != "" && matchRequest.TagLine != "" {
		// Otherwise, look up PUUID using Riot ID
		summoner, err := handler.riotService.GetSummonerByRiotID(matchRequest.Region, matchRequest.GameName, matchRequest.TagLine)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		puuid = summoner.PUUID
	} else {
		http.Error(writer, "either (gameName and tagLine) or puuid is required", http.StatusBadRequest)
		return
	}

	// Set default count if not provided
	count := matchRequest.Count
	if count <= 0 {
		count = 20
	}

	// Get match history using PUUID
	matches, err := handler.riotService.GetMatchHistory(matchRequest.Region, puuid, count)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(matches)
}
