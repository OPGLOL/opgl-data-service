package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OPGLOL/opgl-data/internal/models"
)

// MockRiotService is a mock implementation of RiotServiceInterface for testing
type MockRiotService struct {
	GetSummonerByRiotIDFunc func(region, gameName, tagLine string) (*models.Summoner, error)
	GetSummonerByPUUIDFunc  func(region, puuid string) (*models.Summoner, error)
	GetMatchHistoryFunc     func(region, puuid string, count int) ([]models.Match, error)
	GetMatchDetailsFunc     func(region, matchID string) (*models.Match, error)
}

func (m *MockRiotService) GetSummonerByRiotID(region, gameName, tagLine string) (*models.Summoner, error) {
	if m.GetSummonerByRiotIDFunc != nil {
		return m.GetSummonerByRiotIDFunc(region, gameName, tagLine)
	}
	return nil, nil
}

func (m *MockRiotService) GetSummonerByPUUID(region, puuid string) (*models.Summoner, error) {
	if m.GetSummonerByPUUIDFunc != nil {
		return m.GetSummonerByPUUIDFunc(region, puuid)
	}
	return nil, nil
}

func (m *MockRiotService) GetMatchHistory(region, puuid string, count int) ([]models.Match, error) {
	if m.GetMatchHistoryFunc != nil {
		return m.GetMatchHistoryFunc(region, puuid, count)
	}
	return nil, nil
}

func (m *MockRiotService) GetMatchDetails(region, matchID string) (*models.Match, error) {
	if m.GetMatchDetailsFunc != nil {
		return m.GetMatchDetailsFunc(region, matchID)
	}
	return nil, nil
}

// TestNewHandler tests the NewHandler constructor
func TestNewHandler(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)

	if handler == nil {
		t.Fatal("Expected handler to not be nil")
	}

	if handler.riotService != mockService {
		t.Error("Expected riotService to be set correctly")
	}
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	handler := &Handler{riotService: nil}

	request, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.HealthCheck(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response map[string]string
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}

	if response["service"] != "opgl-data" {
		t.Errorf("Expected service 'opgl-data', got '%s'", response["service"])
	}
}

// TestHealthCheckContentType tests that health check returns JSON content type
func TestHealthCheckContentType(t *testing.T) {
	handler := &Handler{riotService: nil}

	request, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.HealthCheck(responseRecorder, request)

	contentType := responseRecorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestGetSummonerByRiotID_Success tests successful summoner lookup
func TestGetSummonerByRiotID_Success(t *testing.T) {
	expectedSummoner := &models.Summoner{
		ID:            "test-id",
		AccountID:     "test-account-id",
		PUUID:         "test-puuid",
		Name:          "TestPlayer",
		ProfileIconID: 1234,
		SummonerLevel: 100,
	}

	mockService := &MockRiotService{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			if region != "na" || gameName != "TestPlayer" || tagLine != "NA1" {
				t.Errorf("Unexpected parameters: region=%s, gameName=%s, tagLine=%s", region, gameName, tagLine)
			}
			return expectedSummoner, nil
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, err := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetSummonerByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response models.Summoner
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.PUUID != expectedSummoner.PUUID {
		t.Errorf("Expected PUUID '%s', got '%s'", expectedSummoner.PUUID, response.PUUID)
	}
}

// TestGetSummonerByRiotID_InvalidJSON tests invalid JSON request body
func TestGetSummonerByRiotID_InvalidJSON(t *testing.T) {
	handler := NewHandler(&MockRiotService{})

	request, err := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	handler.GetSummonerByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestGetSummonerByRiotID_MissingFields tests missing required fields
func TestGetSummonerByRiotID_MissingFields(t *testing.T) {
	testCases := []struct {
		name        string
		requestBody map[string]string
	}{
		{"missing region", map[string]string{"gameName": "Test", "tagLine": "NA1"}},
		{"missing gameName", map[string]string{"region": "na", "tagLine": "NA1"}},
		{"missing tagLine", map[string]string{"region": "na", "gameName": "Test"}},
		{"empty region", map[string]string{"region": "", "gameName": "Test", "tagLine": "NA1"}},
		{"empty gameName", map[string]string{"region": "na", "gameName": "", "tagLine": "NA1"}},
		{"empty tagLine", map[string]string{"region": "na", "gameName": "Test", "tagLine": ""}},
	}

	handler := NewHandler(&MockRiotService{})

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(testCase.requestBody)
			request, _ := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBuffer(bodyBytes))
			request.Header.Set("Content-Type", "application/json")

			responseRecorder := httptest.NewRecorder()
			handler.GetSummonerByRiotID(responseRecorder, request)

			if responseRecorder.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
			}
		})
	}
}

// TestGetSummonerByRiotID_ServiceError tests service error handling
func TestGetSummonerByRiotID_ServiceError(t *testing.T) {
	mockService := &MockRiotService{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return nil, errors.New("API error")
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]string{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetSummonerByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestGetMatchesByRiotID_Success tests successful match history lookup with Riot ID
func TestGetMatchesByRiotID_Success(t *testing.T) {
	expectedSummoner := &models.Summoner{PUUID: "test-puuid"}
	expectedMatches := []models.Match{
		{MatchID: "NA1_123", GameMode: "CLASSIC"},
		{MatchID: "NA1_124", GameMode: "CLASSIC"},
	}

	mockService := &MockRiotService{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return expectedSummoner, nil
		},
		GetMatchHistoryFunc: func(region, puuid string, count int) ([]models.Match, error) {
			if puuid != expectedSummoner.PUUID {
				t.Errorf("Expected PUUID '%s', got '%s'", expectedSummoner.PUUID, puuid)
			}
			return expectedMatches, nil
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]interface{}{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
		"count":    10,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	var response []models.Match
	err := json.NewDecoder(responseRecorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != len(expectedMatches) {
		t.Errorf("Expected %d matches, got %d", len(expectedMatches), len(response))
	}
}

// TestGetMatchesByRiotID_WithPUUID tests match history lookup with direct PUUID
func TestGetMatchesByRiotID_WithPUUID(t *testing.T) {
	expectedMatches := []models.Match{
		{MatchID: "NA1_123", GameMode: "CLASSIC"},
	}

	mockService := &MockRiotService{
		GetMatchHistoryFunc: func(region, puuid string, count int) ([]models.Match, error) {
			if puuid != "direct-puuid" {
				t.Errorf("Expected PUUID 'direct-puuid', got '%s'", puuid)
			}
			return expectedMatches, nil
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]interface{}{
		"region": "na",
		"puuid":  "direct-puuid",
		"count":  10,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestGetMatchesByRiotID_DefaultCount tests default count when not provided
func TestGetMatchesByRiotID_DefaultCount(t *testing.T) {
	var capturedCount int

	mockService := &MockRiotService{
		GetMatchHistoryFunc: func(region, puuid string, count int) ([]models.Match, error) {
			capturedCount = count
			return []models.Match{}, nil
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]interface{}{
		"region": "na",
		"puuid":  "test-puuid",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if capturedCount != 20 {
		t.Errorf("Expected default count 20, got %d", capturedCount)
	}
}

// TestGetMatchesByRiotID_InvalidJSON tests invalid JSON request body
func TestGetMatchesByRiotID_InvalidJSON(t *testing.T) {
	handler := NewHandler(&MockRiotService{})

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBufferString("invalid json"))

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestGetMatchesByRiotID_MissingRegion tests missing region field
func TestGetMatchesByRiotID_MissingRegion(t *testing.T) {
	handler := NewHandler(&MockRiotService{})

	requestBody := map[string]interface{}{
		"puuid": "test-puuid",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestGetMatchesByRiotID_MissingIdentifiers tests missing both PUUID and Riot ID
func TestGetMatchesByRiotID_MissingIdentifiers(t *testing.T) {
	handler := NewHandler(&MockRiotService{})

	requestBody := map[string]interface{}{
		"region": "na",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestGetMatchesByRiotID_SummonerLookupError tests error during summoner lookup
func TestGetMatchesByRiotID_SummonerLookupError(t *testing.T) {
	mockService := &MockRiotService{
		GetSummonerByRiotIDFunc: func(region, gameName, tagLine string) (*models.Summoner, error) {
			return nil, errors.New("summoner not found")
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]interface{}{
		"region":   "na",
		"gameName": "TestPlayer",
		"tagLine":  "NA1",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestGetMatchesByRiotID_MatchHistoryError tests error during match history lookup
func TestGetMatchesByRiotID_MatchHistoryError(t *testing.T) {
	mockService := &MockRiotService{
		GetMatchHistoryFunc: func(region, puuid string, count int) ([]models.Match, error) {
			return nil, errors.New("match history error")
		},
	}

	handler := NewHandler(mockService)

	requestBody := map[string]interface{}{
		"region": "na",
		"puuid":  "test-puuid",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	request, _ := http.NewRequest("POST", "/api/v1/matches", bytes.NewBuffer(bodyBytes))
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.GetMatchesByRiotID(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}
