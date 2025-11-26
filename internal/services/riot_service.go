package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/OPGLOL/opgl-data-service/internal/models"
)

// RiotService handles all interactions with the Riot Games API
type RiotService struct {
	// Riot Games API key for authentication
	apiKey string
	// HTTP client with configured timeout
	httpClient *http.Client
	// Base URL override for testing (optional)
	baseURLOverride string
}

// NewRiotService creates a new RiotService with the provided API key
func NewRiotService(apiKey string) *RiotService {
	return &RiotService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURLOverride: "",
	}
}

// NewRiotServiceWithBaseURL creates a RiotService with a custom base URL (for testing)
func NewRiotServiceWithBaseURL(apiKey string, baseURL string, httpClient *http.Client) *RiotService {
	return &RiotService{
		apiKey:          apiKey,
		httpClient:      httpClient,
		baseURLOverride: baseURL,
	}
}

// getRegionalURL returns the correct API URL based on region
func (riotService *RiotService) getRegionalURL(region string) string {
	// Map region codes to Riot API regional routing values
	regionalRouting := map[string]string{
		"na":  "na1.api.riotgames.com",
		"euw": "euw1.api.riotgames.com",
		"eune": "eun1.api.riotgames.com",
		"kr":  "kr.api.riotgames.com",
		"br":  "br1.api.riotgames.com",
		"jp":  "jp1.api.riotgames.com",
		"ru":  "ru.api.riotgames.com",
		"oce": "oc1.api.riotgames.com",
		"tr":  "tr1.api.riotgames.com",
		"lan": "la1.api.riotgames.com",
		"las": "la2.api.riotgames.com",
	}

	if url, exists := regionalRouting[region]; exists {
		return url
	}

	// Default to NA if region is not recognized
	return regionalRouting["na"]
}

// getMatchRegionalURL returns the correct match API URL based on region
func (riotService *RiotService) getMatchRegionalURL(region string) string {
	// Match API uses continental routing
	continentalRouting := map[string]string{
		"na":  "americas.api.riotgames.com",
		"br":  "americas.api.riotgames.com",
		"lan": "americas.api.riotgames.com",
		"las": "americas.api.riotgames.com",
		"euw": "europe.api.riotgames.com",
		"eune": "europe.api.riotgames.com",
		"tr":  "europe.api.riotgames.com",
		"ru":  "europe.api.riotgames.com",
		"kr":  "asia.api.riotgames.com",
		"jp":  "asia.api.riotgames.com",
		"oce": "sea.api.riotgames.com",
	}

	if url, exists := continentalRouting[region]; exists {
		return url
	}

	// Default to Americas if region is not recognized
	return continentalRouting["na"]
}

// makeRequest performs an HTTP GET request to the Riot API
func (riotService *RiotService) makeRequest(url string, target interface{}) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key to request header
	request.Header.Add("X-Riot-Token", riotService.apiKey)

	response, err := riotService.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(body))
	}

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// buildURL creates the full URL, using baseURLOverride if set (for testing)
func (riotService *RiotService) buildURL(baseURL string, path string) string {
	if riotService.baseURLOverride != "" {
		return riotService.baseURLOverride + path
	}
	return fmt.Sprintf("https://%s%s", baseURL, path)
}

// GetSummonerByRiotID retrieves summoner information using Riot ID (gameName#tagLine)
// This is the new Riot API method that replaced the deprecated by-name endpoint
func (riotService *RiotService) GetSummonerByRiotID(region string, gameName string, tagLine string) (*models.Summoner, error) {
	// Step 1: Get account info (PUUID) using Riot Account API
	accountURL := riotService.getMatchRegionalURL(region)
	accountPath := fmt.Sprintf("/riot/account/v1/accounts/by-riot-id/%s/%s", gameName, tagLine)
	accountEndpoint := riotService.buildURL(accountURL, accountPath)

	var accountInfo struct {
		PUUID    string `json:"puuid"`
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	if err := riotService.makeRequest(accountEndpoint, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// Step 2: Get summoner details using PUUID
	return riotService.GetSummonerByPUUID(region, accountInfo.PUUID)
}

// GetSummonerByPUUID retrieves summoner information by PUUID
func (riotService *RiotService) GetSummonerByPUUID(region string, puuid string) (*models.Summoner, error) {
	baseURL := riotService.getRegionalURL(region)
	path := fmt.Sprintf("/lol/summoner/v4/summoners/by-puuid/%s", puuid)
	url := riotService.buildURL(baseURL, path)

	var summoner models.Summoner
	if err := riotService.makeRequest(url, &summoner); err != nil {
		return nil, fmt.Errorf("failed to get summoner: %w", err)
	}

	return &summoner, nil
}

// GetMatchHistory retrieves recent match IDs for a player and fetches full match details
func (riotService *RiotService) GetMatchHistory(region string, puuid string, count int) ([]models.Match, error) {
	baseURL := riotService.getMatchRegionalURL(region)
	path := fmt.Sprintf("/lol/match/v5/matches/by-puuid/%s/ids?start=0&count=%d", puuid, count)
	matchListURL := riotService.buildURL(baseURL, path)

	var matchIDs []string
	if err := riotService.makeRequest(matchListURL, &matchIDs); err != nil {
		return nil, fmt.Errorf("failed to get match list: %w", err)
	}

	// Fetch details for each match
	matches := make([]models.Match, 0, len(matchIDs))
	for _, matchID := range matchIDs {
		match, err := riotService.GetMatchDetails(region, matchID)
		if err != nil {
			// Log error but continue processing other matches
			continue
		}
		matches = append(matches, *match)
	}

	return matches, nil
}

// GetMatchDetails retrieves detailed information for a specific match
func (riotService *RiotService) GetMatchDetails(region string, matchID string) (*models.Match, error) {
	baseURL := riotService.getMatchRegionalURL(region)
	path := fmt.Sprintf("/lol/match/v5/matches/%s", matchID)
	url := riotService.buildURL(baseURL, path)

	var rawMatch struct {
		Metadata struct {
			MatchID string `json:"matchId"`
		} `json:"metadata"`
		Info struct {
			GameCreation int64  `json:"gameCreation"`
			GameDuration int    `json:"gameDuration"`
			GameMode     string `json:"gameMode"`
			GameType     string `json:"gameType"`
			Participants []struct {
				PUUID                       string `json:"puuid"`
				SummonerName                string `json:"summonerName"`
				ChampionID                  int    `json:"championId"`
				ChampionName                string `json:"championName"`
				Kills                       int    `json:"kills"`
				Deaths                      int    `json:"deaths"`
				Assists                     int    `json:"assists"`
				GoldEarned                  int    `json:"goldEarned"`
				TotalDamageDealtToChampions int    `json:"totalDamageDealtToChampions"`
				TotalDamageTaken            int    `json:"totalDamageTaken"`
				VisionScore                 int    `json:"visionScore"`
				TotalMinionsKilled          int    `json:"totalMinionsKilled"`
				Win                         bool   `json:"win"`
				TeamPosition                string `json:"teamPosition"`
			} `json:"participants"`
		} `json:"info"`
	}

	if err := riotService.makeRequest(url, &rawMatch); err != nil {
		return nil, fmt.Errorf("failed to get match details: %w", err)
	}

	// Convert raw match data to our model
	match := &models.Match{
		MatchID:      rawMatch.Metadata.MatchID,
		GameCreation: time.UnixMilli(rawMatch.Info.GameCreation),
		GameDuration: rawMatch.Info.GameDuration,
		GameMode:     rawMatch.Info.GameMode,
		GameType:     rawMatch.Info.GameType,
		Participants: make([]models.Participant, len(rawMatch.Info.Participants)),
	}

	for i, participant := range rawMatch.Info.Participants {
		match.Participants[i] = models.Participant{
			PUUID:                       participant.PUUID,
			SummonerName:                participant.SummonerName,
			ChampionID:                  participant.ChampionID,
			ChampionName:                participant.ChampionName,
			Kills:                       participant.Kills,
			Deaths:                      participant.Deaths,
			Assists:                     participant.Assists,
			GoldEarned:                  participant.GoldEarned,
			TotalDamageDealtToChampions: participant.TotalDamageDealtToChampions,
			TotalDamageTaken:            participant.TotalDamageTaken,
			VisionScore:                 participant.VisionScore,
			TotalMinionsKilled:          participant.TotalMinionsKilled,
			Win:                         participant.Win,
			TeamPosition:                participant.TeamPosition,
		}
	}

	return match, nil
}
