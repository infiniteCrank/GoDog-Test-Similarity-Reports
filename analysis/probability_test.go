package analysis

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	messages "github.com/cucumber/messages/go/v22"
)

func TestDetermineComponentProbability(t *testing.T) {
	tests := []struct {
		keywords map[string]int
		expected float64
	}{
		{map[string]int{"button": 1}, 0.8},
		{map[string]int{"button": 0}, 0.2},
	}

	for _, test := range tests {
		result := determineComponentProbability(test.keywords)
		if result != test.expected {
			t.Errorf("Expected %f, got %f", test.expected, result)
		}
	}
}

// Other probability tests similar to the above...

func TestHandleGherkin(t *testing.T) {
	// Create a sample Gherkin input
	input := "Feature: User login\n  Scenario: User can log in successfully\n  Scenario: API responds with a 200 status"
	req := httptest.NewRequest("POST", "/analyze", strings.NewReader(input))
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder() // Create a ResponseRecorder to capture the response

	// Create a multipart form with Gherkin input
	body := &strings.Builder{}
	mw := multipart.NewWriter(body)
	fw, _ := mw.CreateFormFile("files", "test.feature")
	fw.Write([]byte(input))
	mw.Close()

	req.Body = io.NopCloser(strings.NewReader(body.String()))
	req.Header.Set("Content-Type", mw.FormDataContentType())

	// Call the HandleGherkin function
	HandleGherkin(w, req)

	// Check the response status code
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}

	// Decode the response body
	var result []ScenarioProbability
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	// Assert the length of result
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	// Further checks on the probabilities returned
	for _, scenario := range result {
		if scenario.Probability["component"] < 0 || scenario.Probability["component"] > 1 {
			t.Errorf("Component probability out of bounds for scenario '%s'", scenario.ScenarioName)
		}
		if scenario.Probability["integration"] < 0 || scenario.Probability["integration"] > 1 {
			t.Errorf("Integration probability out of bounds for scenario '%s'", scenario.ScenarioName)
		}
		if scenario.Probability["end_to_end"] < 0 || scenario.Probability["end_to_end"] > 1 {
			t.Errorf("End-to-end probability out of bounds for scenario '%s'", scenario.ScenarioName)
		}
		if scenario.Probability["regression"] < 0 || scenario.Probability["regression"] > 1 {
			t.Errorf("Regression probability out of bounds for scenario '%s'", scenario.ScenarioName)
		}
	}
}

func TestAnalyzeGherkinDocument(t *testing.T) {
	input := "Feature: Sample Feature\n  Scenario: Sample scenario with user login"
	reader := strings.NewReader(input)
	uuid := &messages.UUID{}

	results, err := analyzeGherkinDocument(reader, uuid)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 scenario, got %d", len(results))
	}

	if results[0].ScenarioName != "Sample scenario with user login" {
		t.Errorf("Expected scenario name '%s', got '%s'", "Sample scenario with user login", results[0].ScenarioName)
	}

	if results[0].Probability["component"] < 0 || results[0].Probability["component"] > 1 {
		t.Errorf("Component probability out of bounds: %f", results[0].Probability["component"])
	}
}
