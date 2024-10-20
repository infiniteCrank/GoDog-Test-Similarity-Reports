package analysis

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestE2EHandleGherkin(t *testing.T) {
	// Given a Gherkin feature file content
	input := "Feature: User login\n  Scenario: User can log in successfully\n  Scenario: API responds with a 200 status"

	// Prepare a multipart form for file upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", "test.feature")
	if err != nil {
		t.Fatalf("Error creating form file: %v", err)
	}
	io.Copy(part, strings.NewReader(input))
	writer.Close()

	// Create a new HTTP request
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Record the response
	recorder := httptest.NewRecorder()

	// Call the actual handler
	HandleGherkin(recorder, req)

	// Check the status code
	res := recorder.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", res.StatusCode)
	}

	// Check the response Content-Type
	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Expected Content-Type 'application/json', got %s", contentType)
	}

	// Decode the response body
	var result []ScenarioProbability
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	// Assert that we have results
	if len(result) != 2 {
		t.Fatalf("Expected 2 scenarios, got %d", len(result))
	}
}
