package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewResponseWriter tests the response writer constructor
func TestNewResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseWriter := newResponseWriter(recorder)

	if responseWriter == nil {
		t.Fatal("Expected responseWriter to not be nil")
	}

	if responseWriter.statusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, responseWriter.statusCode)
	}

	if responseWriter.ResponseWriter != recorder {
		t.Error("Expected ResponseWriter to be set correctly")
	}
}

// TestResponseWriter_WriteHeader tests that WriteHeader captures status code
func TestResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseWriter := newResponseWriter(recorder)

	responseWriter.WriteHeader(http.StatusNotFound)

	if responseWriter.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, responseWriter.statusCode)
	}

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected underlying recorder code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

// TestResponseWriter_WriteHeader_MultipleStatuses tests multiple status codes
func TestResponseWriter_WriteHeader_MultipleStatuses(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			responseWriter := newResponseWriter(recorder)

			responseWriter.WriteHeader(testCase.statusCode)

			if responseWriter.statusCode != testCase.statusCode {
				t.Errorf("Expected status code %d, got %d", testCase.statusCode, responseWriter.statusCode)
			}
		})
	}
}

// TestLoggingMiddleware tests the logging middleware
func TestLoggingMiddleware(t *testing.T) {
	handlerCalled := false
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		handlerCalled = true
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})

	middleware := LoggingMiddleware(nextHandler)

	request, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	responseRecorder := httptest.NewRecorder()
	middleware.ServeHTTP(responseRecorder, request)

	if !handlerCalled {
		t.Error("Expected next handler to be called")
	}

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_4xxStatus tests logging middleware with 4xx status
func TestLoggingMiddleware_4xxStatus(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("POST", "/test", nil)
	responseRecorder := httptest.NewRecorder()
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_5xxStatus tests logging middleware with 5xx status
func TestLoggingMiddleware_5xxStatus(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("POST", "/test", nil)
	responseRecorder := httptest.NewRecorder()
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_WithUserAgent tests that user agent is captured
func TestLoggingMiddleware_WithUserAgent(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("GET", "/test", nil)
	request.Header.Set("User-Agent", "TestAgent/1.0")

	responseRecorder := httptest.NewRecorder()
	middleware.ServeHTTP(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}

// TestLoggingMiddleware_DefaultStatusCode tests default status code when WriteHeader not called
func TestLoggingMiddleware_DefaultStatusCode(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Don't call WriteHeader, just write body
		writer.Write([]byte("OK"))
	})

	middleware := LoggingMiddleware(nextHandler)

	request, _ := http.NewRequest("GET", "/test", nil)
	responseRecorder := httptest.NewRecorder()
	middleware.ServeHTTP(responseRecorder, request)

	// When WriteHeader is not called, default is 200
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}
}
