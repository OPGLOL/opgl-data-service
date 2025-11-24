package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSetupRouter tests that the router is set up correctly
func TestSetupRouter(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)
	router := SetupRouter(handler)

	if router == nil {
		t.Fatal("Expected router to not be nil")
	}
}

// TestSetupRouter_HealthEndpoint tests the health endpoint is registered
func TestSetupRouter_HealthEndpoint(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)
	router := SetupRouter(handler)

	request, err := http.NewRequest("POST", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestSetupRouter_HealthEndpoint_WrongMethod tests health endpoint rejects GET
func TestSetupRouter_HealthEndpoint_WrongMethod(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)
	router := SetupRouter(handler)

	request, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, responseRecorder.Code)
	}
}

// TestSetupRouter_SummonerEndpoint tests the summoner endpoint is registered
func TestSetupRouter_SummonerEndpoint(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)
	router := SetupRouter(handler)

	// Send empty JSON body to avoid nil pointer
	request, err := http.NewRequest("POST", "/api/v1/summoner", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	// Should get 400 because required fields missing, not 404
	if responseRecorder.Code == http.StatusNotFound {
		t.Error("Summoner endpoint not found - route not registered")
	}
}

// TestSetupRouter_MatchesEndpoint tests the matches endpoint is registered
func TestSetupRouter_MatchesEndpoint(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)
	router := SetupRouter(handler)

	// Send empty JSON body to avoid nil pointer
	request, err := http.NewRequest("POST", "/api/v1/matches", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	// Should get 400 because required fields missing, not 404
	if responseRecorder.Code == http.StatusNotFound {
		t.Error("Matches endpoint not found - route not registered")
	}
}

// TestSetupRouter_NotFoundEndpoint tests unknown endpoints return 404
func TestSetupRouter_NotFoundEndpoint(t *testing.T) {
	mockService := &MockRiotService{}
	handler := NewHandler(mockService)
	router := SetupRouter(handler)

	request, err := http.NewRequest("POST", "/api/v1/unknown", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}
}
