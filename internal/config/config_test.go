package config

import (
	"os"
	"testing"
)

// TestLoadConfig_DefaultValues tests that default values are set correctly
func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("RIOT_API_KEY")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	config := LoadConfig()

	if config == nil {
		t.Fatal("Expected config to not be nil")
	}

	if config.ServerPort != "8081" {
		t.Errorf("Expected default ServerPort '8081', got '%s'", config.ServerPort)
	}

	if config.RiotAPIKey != "" {
		t.Errorf("Expected empty RiotAPIKey, got '%s'", config.RiotAPIKey)
	}

	if config.DatabaseURL != "" {
		t.Errorf("Expected empty DatabaseURL, got '%s'", config.DatabaseURL)
	}
}

// TestLoadConfig_WithEnvironmentVariables tests loading from environment
func TestLoadConfig_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("RIOT_API_KEY", "test-api-key")
	os.Setenv("PORT", "9000")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test")

	// Clean up after test
	defer func() {
		os.Unsetenv("RIOT_API_KEY")
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
	}()

	config := LoadConfig()

	if config.RiotAPIKey != "test-api-key" {
		t.Errorf("Expected RiotAPIKey 'test-api-key', got '%s'", config.RiotAPIKey)
	}

	if config.ServerPort != "9000" {
		t.Errorf("Expected ServerPort '9000', got '%s'", config.ServerPort)
	}

	if config.DatabaseURL != "postgres://localhost:5432/test" {
		t.Errorf("Expected DatabaseURL 'postgres://localhost:5432/test', got '%s'", config.DatabaseURL)
	}
}

// TestLoadConfig_PartialEnvironment tests loading with some env vars set
func TestLoadConfig_PartialEnvironment(t *testing.T) {
	// Set only some environment variables
	os.Setenv("RIOT_API_KEY", "partial-api-key")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	defer func() {
		os.Unsetenv("RIOT_API_KEY")
	}()

	config := LoadConfig()

	if config.RiotAPIKey != "partial-api-key" {
		t.Errorf("Expected RiotAPIKey 'partial-api-key', got '%s'", config.RiotAPIKey)
	}

	if config.ServerPort != "8081" {
		t.Errorf("Expected default ServerPort '8081', got '%s'", config.ServerPort)
	}

	if config.DatabaseURL != "" {
		t.Errorf("Expected empty DatabaseURL, got '%s'", config.DatabaseURL)
	}
}

// TestLoadConfig_EmptyPort tests that empty PORT defaults to 8081
func TestLoadConfig_EmptyPort(t *testing.T) {
	os.Setenv("PORT", "")
	defer os.Unsetenv("PORT")

	config := LoadConfig()

	if config.ServerPort != "8081" {
		t.Errorf("Expected ServerPort '8081' for empty PORT, got '%s'", config.ServerPort)
	}
}

// TestLoadConfig_CustomPort tests custom port setting
func TestLoadConfig_CustomPort(t *testing.T) {
	testPorts := []string{"3000", "8080", "9999", "80"}

	for _, port := range testPorts {
		t.Run("Port_"+port, func(t *testing.T) {
			os.Setenv("PORT", port)
			defer os.Unsetenv("PORT")

			config := LoadConfig()

			if config.ServerPort != port {
				t.Errorf("Expected ServerPort '%s', got '%s'", port, config.ServerPort)
			}
		})
	}
}

// TestConfigStruct tests the Config struct fields
func TestConfigStruct(t *testing.T) {
	config := &Config{
		RiotAPIKey:  "test-key",
		ServerPort:  "8081",
		DatabaseURL: "test-url",
	}

	if config.RiotAPIKey != "test-key" {
		t.Errorf("Expected RiotAPIKey 'test-key', got '%s'", config.RiotAPIKey)
	}

	if config.ServerPort != "8081" {
		t.Errorf("Expected ServerPort '8081', got '%s'", config.ServerPort)
	}

	if config.DatabaseURL != "test-url" {
		t.Errorf("Expected DatabaseURL 'test-url', got '%s'", config.DatabaseURL)
	}
}
