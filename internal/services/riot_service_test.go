package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewRiotService tests the RiotService constructor
func TestNewRiotService(t *testing.T) {
	service := NewRiotService("test-api-key")

	if service == nil {
		t.Fatal("Expected service to not be nil")
	}

	if service.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got '%s'", service.apiKey)
	}

	if service.httpClient == nil {
		t.Error("Expected httpClient to not be nil")
	}
}

// TestGetRegionalURL tests regional URL mapping
func TestGetRegionalURL(t *testing.T) {
	service := NewRiotService("test-key")

	testCases := []struct {
		region      string
		expectedURL string
	}{
		{"na", "na1.api.riotgames.com"},
		{"euw", "euw1.api.riotgames.com"},
		{"eune", "eun1.api.riotgames.com"},
		{"kr", "kr.api.riotgames.com"},
		{"br", "br1.api.riotgames.com"},
		{"jp", "jp1.api.riotgames.com"},
		{"ru", "ru.api.riotgames.com"},
		{"oce", "oc1.api.riotgames.com"},
		{"tr", "tr1.api.riotgames.com"},
		{"lan", "la1.api.riotgames.com"},
		{"las", "la2.api.riotgames.com"},
		{"unknown", "na1.api.riotgames.com"}, // Default to NA
	}

	for _, testCase := range testCases {
		t.Run(testCase.region, func(t *testing.T) {
			url := service.getRegionalURL(testCase.region)
			if url != testCase.expectedURL {
				t.Errorf("Expected URL '%s' for region '%s', got '%s'", testCase.expectedURL, testCase.region, url)
			}
		})
	}
}

// TestGetMatchRegionalURL tests continental URL mapping for match API
func TestGetMatchRegionalURL(t *testing.T) {
	service := NewRiotService("test-key")

	testCases := []struct {
		region      string
		expectedURL string
	}{
		{"na", "americas.api.riotgames.com"},
		{"br", "americas.api.riotgames.com"},
		{"lan", "americas.api.riotgames.com"},
		{"las", "americas.api.riotgames.com"},
		{"euw", "europe.api.riotgames.com"},
		{"eune", "europe.api.riotgames.com"},
		{"tr", "europe.api.riotgames.com"},
		{"ru", "europe.api.riotgames.com"},
		{"kr", "asia.api.riotgames.com"},
		{"jp", "asia.api.riotgames.com"},
		{"oce", "sea.api.riotgames.com"},
		{"unknown", "americas.api.riotgames.com"}, // Default to Americas
	}

	for _, testCase := range testCases {
		t.Run(testCase.region, func(t *testing.T) {
			url := service.getMatchRegionalURL(testCase.region)
			if url != testCase.expectedURL {
				t.Errorf("Expected URL '%s' for region '%s', got '%s'", testCase.expectedURL, testCase.region, url)
			}
		})
	}
}

// TestMakeRequest_Success tests successful API request
func TestMakeRequest_Success(t *testing.T) {
	expectedResponse := map[string]string{"status": "ok"}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Verify API key header
		apiKey := request.Header.Get("X-Riot-Token")
		if apiKey != "test-api-key" {
			t.Errorf("Expected API key 'test-api-key', got '%s'", apiKey)
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(expectedResponse)
	}))
	defer server.Close()

	service := NewRiotService("test-api-key")

	var result map[string]string
	err := service.makeRequest(server.URL, &result)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", result["status"])
	}
}

// TestMakeRequest_NonOKStatus tests handling of non-200 status codes
func TestMakeRequest_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Not found"))
	}))
	defer server.Close()

	service := NewRiotService("test-api-key")

	var result map[string]string
	err := service.makeRequest(server.URL, &result)

	if err == nil {
		t.Fatal("Expected error for non-OK status")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error to contain '404', got: %v", err)
	}
}

// TestMakeRequest_InvalidJSON tests handling of invalid JSON response
func TestMakeRequest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Write([]byte("invalid json"))
	}))
	defer server.Close()

	service := NewRiotService("test-api-key")

	var result map[string]string
	err := service.makeRequest(server.URL, &result)

	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("Expected error about decoding, got: %v", err)
	}
}

// TestMakeRequest_InvalidURL tests handling of invalid URL
func TestMakeRequest_InvalidURL(t *testing.T) {
	service := NewRiotService("test-api-key")

	var result map[string]string
	err := service.makeRequest("http://invalid-url-that-will-fail:99999", &result)

	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}
}

// TestGetMatchHistory_EmptyMatches tests empty match list
func TestGetMatchHistory_EmptyMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if strings.Contains(request.URL.Path, "/ids") {
			json.NewEncoder(writer).Encode([]string{})
		}
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	matches, err := service.GetMatchHistory("na", "test-puuid", 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

// TestRiotServiceInterface_Implementation verifies interface implementation
func TestRiotServiceInterface_Implementation(t *testing.T) {
	service := NewRiotService("test-key")

	// Verify that RiotService implements RiotServiceInterface
	var _ RiotServiceInterface = service

	t.Log("RiotService correctly implements RiotServiceInterface")
}

// TestNewRiotServiceWithBaseURL tests the constructor with base URL override
func TestNewRiotServiceWithBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-key", server.URL, server.Client())

	if service.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got '%s'", service.apiKey)
	}

	if service.baseURLOverride != server.URL {
		t.Errorf("Expected baseURLOverride '%s', got '%s'", server.URL, service.baseURLOverride)
	}
}

// TestBuildURL tests the URL building function
func TestBuildURL(t *testing.T) {
	service := NewRiotService("test-key")

	// Test without override
	url := service.buildURL("na1.api.riotgames.com", "/test/path")
	if url != "https://na1.api.riotgames.com/test/path" {
		t.Errorf("Expected 'https://na1.api.riotgames.com/test/path', got '%s'", url)
	}

	// Test with override
	service.baseURLOverride = "http://localhost:8080"
	url = service.buildURL("na1.api.riotgames.com", "/test/path")
	if url != "http://localhost:8080/test/path" {
		t.Errorf("Expected 'http://localhost:8080/test/path', got '%s'", url)
	}
}

// TestGetSummonerByRiotID_Success tests the full GetSummonerByRiotID flow
func TestGetSummonerByRiotID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		if strings.Contains(request.URL.Path, "riot/account") {
			json.NewEncoder(writer).Encode(map[string]interface{}{
				"puuid":    "test-puuid-123",
				"gameName": "TestPlayer",
				"tagLine":  "NA1",
			})
		} else if strings.Contains(request.URL.Path, "summoners/by-puuid") {
			json.NewEncoder(writer).Encode(map[string]interface{}{
				"id":            "summoner-id",
				"accountId":     "account-id",
				"puuid":         "test-puuid-123",
				"name":          "TestPlayer",
				"profileIconId": 1234,
				"summonerLevel": 150,
			})
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	summoner, err := service.GetSummonerByRiotID("na", "TestPlayer", "NA1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if summoner.PUUID != "test-puuid-123" {
		t.Errorf("Expected PUUID 'test-puuid-123', got '%s'", summoner.PUUID)
	}
}

// TestGetSummonerByRiotID_AccountError tests error handling for account lookup
func TestGetSummonerByRiotID_AccountError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Account not found"))
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	_, err := service.GetSummonerByRiotID("na", "NonExistent", "NA1")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestGetSummonerByPUUID_Success tests successful summoner lookup by PUUID
func TestGetSummonerByPUUID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]interface{}{
			"id":            "summoner-id",
			"accountId":     "account-id",
			"puuid":         "test-puuid-123",
			"name":          "TestPlayer",
			"profileIconId": 1234,
			"summonerLevel": 150,
		})
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	summoner, err := service.GetSummonerByPUUID("na", "test-puuid-123")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if summoner.PUUID != "test-puuid-123" {
		t.Errorf("Expected PUUID 'test-puuid-123', got '%s'", summoner.PUUID)
	}
}

// TestGetSummonerByPUUID_Error tests error handling for summoner lookup
func TestGetSummonerByPUUID_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Summoner not found"))
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	_, err := service.GetSummonerByPUUID("na", "invalid-puuid")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestGetMatchHistory_Success tests successful match history retrieval
func TestGetMatchHistory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		if strings.Contains(request.URL.Path, "/ids") {
			json.NewEncoder(writer).Encode([]string{"NA1_123", "NA1_124"})
		} else if strings.Contains(request.URL.Path, "matches/NA1_") {
			json.NewEncoder(writer).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{
					"matchId": "NA1_123",
				},
				"info": map[string]interface{}{
					"gameCreation": 1700000000000,
					"gameDuration": 1800,
					"gameMode":     "CLASSIC",
					"gameType":     "MATCHED_GAME",
					"participants": []map[string]interface{}{},
				},
			})
		}
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	matches, err := service.GetMatchHistory("na", "test-puuid", 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
}

// TestGetMatchHistory_Error tests error handling for match history
func TestGetMatchHistory_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Server error"))
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	_, err := service.GetMatchHistory("na", "test-puuid", 10)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestGetMatchHistory_PartialFailure tests handling of partial match detail failures
func TestGetMatchHistory_PartialFailure(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		if strings.Contains(request.URL.Path, "/ids") {
			json.NewEncoder(writer).Encode([]string{"NA1_123", "NA1_124"})
		} else if strings.Contains(request.URL.Path, "matches/NA1_123") {
			json.NewEncoder(writer).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{"matchId": "NA1_123"},
				"info": map[string]interface{}{
					"gameCreation": 1700000000000,
					"gameDuration": 1800,
					"gameMode":     "CLASSIC",
					"gameType":     "MATCHED_GAME",
					"participants": []map[string]interface{}{},
				},
			})
		} else {
			requestCount++
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	matches, err := service.GetMatchHistory("na", "test-puuid", 10)
	if err != nil {
		t.Fatalf("Expected no error even with partial failures, got: %v", err)
	}

	// Should only have 1 match since the second one failed
	if len(matches) != 1 {
		t.Errorf("Expected 1 match (partial success), got %d", len(matches))
	}
}

// TestGetMatchDetails_Success tests successful match details retrieval
func TestGetMatchDetails_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]interface{}{
			"metadata": map[string]interface{}{
				"matchId": "NA1_123",
			},
			"info": map[string]interface{}{
				"gameCreation": 1700000000000,
				"gameDuration": 1800,
				"gameMode":     "CLASSIC",
				"gameType":     "MATCHED_GAME",
				"participants": []map[string]interface{}{
					{
						"puuid":                       "test-puuid",
						"summonerName":                "TestPlayer",
						"championId":                  103,
						"championName":                "Ahri",
						"kills":                       10,
						"deaths":                      5,
						"assists":                     15,
						"goldEarned":                  15000,
						"totalDamageDealtToChampions": 25000,
						"totalDamageTaken":            18000,
						"visionScore":                 30,
						"totalMinionsKilled":          180,
						"win":                         true,
						"teamPosition":                "MIDDLE",
					},
				},
			},
		})
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	match, err := service.GetMatchDetails("na", "NA1_123")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if match.MatchID != "NA1_123" {
		t.Errorf("Expected matchId 'NA1_123', got '%s'", match.MatchID)
	}

	if len(match.Participants) != 1 {
		t.Errorf("Expected 1 participant, got %d", len(match.Participants))
	}

	if match.Participants[0].ChampionName != "Ahri" {
		t.Errorf("Expected champion 'Ahri', got '%s'", match.Participants[0].ChampionName)
	}
}

// TestGetMatchDetails_Error tests error handling for match details
func TestGetMatchDetails_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Match not found"))
	}))
	defer server.Close()

	service := NewRiotServiceWithBaseURL("test-api-key", server.URL, server.Client())

	_, err := service.GetMatchDetails("na", "invalid-match-id")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// TestMakeRequest_ServerError tests handling of 5xx errors
func TestMakeRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	service := NewRiotService("test-api-key")

	var result map[string]string
	err := service.makeRequest(server.URL, &result)

	if err == nil {
		t.Fatal("Expected error for server error")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to contain '500', got: %v", err)
	}
}

// TestMakeRequest_Unauthorized tests handling of 401 errors
func TestMakeRequest_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	service := NewRiotService("invalid-api-key")

	var result map[string]string
	err := service.makeRequest(server.URL, &result)

	if err == nil {
		t.Fatal("Expected error for unauthorized")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Expected error to contain '401', got: %v", err)
	}
}

// TestMakeRequest_RateLimited tests handling of 429 rate limit errors
func TestMakeRequest_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
		writer.Write([]byte("Rate limit exceeded"))
	}))
	defer server.Close()

	service := NewRiotService("test-api-key")

	var result map[string]string
	err := service.makeRequest(server.URL, &result)

	if err == nil {
		t.Fatal("Expected error for rate limit")
	}

	if !strings.Contains(err.Error(), "429") {
		t.Errorf("Expected error to contain '429', got: %v", err)
	}
}
