package analysis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetermineComponentProbability(t *testing.T) {
	tests := []struct {
		keywords map[string]int
		expected float64
	}{
		{map[string]int{"button": 1}, 0.8}, // Test with a "button" keyword present
		{map[string]int{"button": 0}, 0.2}, // Test without a "button" keyword
	}

	for _, test := range tests {
		result := determineComponentProbability(test.keywords)
		if result != test.expected {
			t.Errorf("Expected %f, got %f", test.expected, result)
		}
	}
}

func TestDetermineIntegrationProbability(t *testing.T) {
	tests := []struct {
		keywords map[string]int
		expected float64
	}{
		{map[string]int{"API": 1}, 0.7},     // Test with "API" present
		{map[string]int{"connect": 1}, 0.7}, // Test with "connect" present
		{map[string]int{}, 0.3},             // Test with no keywords
	}

	for _, test := range tests {
		result := determineIntegrationProbability(test.keywords)
		if result != test.expected {
			t.Errorf("Expected %f, got %f", test.expected, result)
		}
	}
}

func TestDetermineEndToEndProbability(t *testing.T) {
	tests := []struct {
		keywords map[string]int
		expected float64
	}{
		{map[string]int{"user": 1, "login": 1}, 0.6}, // Both "user" and "login" present
		{map[string]int{"user": 1}, 0.2},             // Only "user" present
		{map[string]int{}, 0.2},                      // No keywords present
	}

	for _, test := range tests {
		result := determineEndToEndProbability(test.keywords)
		if result != test.expected {
			t.Errorf("Expected %f, got %f", test.expected, result)
		}
	}
}

func TestDetermineRegressionProbability(t *testing.T) {
	tests := []struct {
		keywords map[string]int
		expected float64
	}{
		{map[string]int{"previous": 1}, 0.5}, // Test with "previous" present
		{map[string]int{}, 0.1},              // No keywords present
	}

	for _, test := range tests {
		result := determineRegressionProbability(test.keywords)
		if result != test.expected {
			t.Errorf("Expected %f, got %f", test.expected, result)
		}
	}
}

func TestHandleGherkin(t *testing.T) {
	// Create a sample Gherkin input
	input := "Feature: User login\n  Scenario: User can log in successfully\n  Scenario: API responds with a 200 status"

	// Create a new HTTP request
	req := httptest.NewRequest("POST", "/analyze", strings.NewReader(input))
	w := httptest.NewRecorder() // Create a ResponseRecorder to capture the response

	// Call the HandleGherkin function
	HandleGherkin(w, req)

	// Check the response status code
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", res.StatusCode)
	}

	// Check the response Content-Type
	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}

	// Optionally, check the response body if needed (JSON decoding may be required here)
}
