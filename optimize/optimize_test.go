package optimize

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// Test for validating feature name structure
func TestValidateFeatureName(t *testing.T) {
	testCases := []struct {
		featureName string
		valid       bool
	}{
		{"Valid Feature", true},
		{"", false},
	}

	for _, tc := range testCases {
		err := validateFeatureName(tc.featureName)
		if (err == nil) != tc.valid {
			t.Errorf("Expected validity: %v, but got error: %v", tc.valid, err)
		}
	}
}

// Test for scenario naming validation
func TestValidateScenarioNames(t *testing.T) {
	scenarios := []Scenario{
		{Name: "Valid Scenario", Steps: []string{"Given I have a valid username"}},
		{Name: "Invalid", Steps: []string{"Given I have a valid username"}},
		{Name: "", Steps: []string{"Given I have a valid username"}},
	}

	issues := validateScenarioNames(scenarios)

	if len(issues) != 2 {
		t.Errorf("Expected 2 naming issue messages, got %d", len(issues))
	}
}

// Integration test for OptimizeFeatureHandler
func TestOptimizeFeatureHandler(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/optimize", OptimizeFeatureHandler).Methods("POST")

	featureContent := `Feature: User login
    Scenario: Successful login
        Given I have a valid username "user1"
        When I perform the login action
        Then I should see a welcome message`

	body := &bytes.Buffer{}
	body.WriteString("feature_file=" + featureContent)
	body.WriteString("&check_naming=true") // Engage naming check

	req := httptest.NewRequest("POST", "/optimize", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if the response body contains expected content
	var res OptimizeResponse
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check if naming issues are included in the response
	if len(res.NamingIssues) != 0 {
		t.Error("Expected no naming issues in the response, but found some.")
	}
}

// Integration test for OptimizeFeatureHandler with naming convention checks
func TestOptimizeFeatureHandlerWithNamingConventionCheck(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/optimize", OptimizeFeatureHandler).Methods("POST")

	// Prepare a sample feature content with naming issues
	featureContent := `Feature: User login
    Scenario: Successful login
        Given I have a valid username "user1"
        When I perform the login action
        Then I should see a welcome message

    Scenario: Invalid
        Given I have a valid username "user2"
        When I perform the login action
        Then I should see a welcome message

    Scenario: 
        Given I have a valid username "user3"
        When I perform the login action
        Then I should see a welcome message`

	// Create a new request with the feature content
	body := &bytes.Buffer{}
	body.WriteString("feature_file=" + featureContent)
	body.WriteString("&check_naming=true") // Engage naming check

	req := httptest.NewRequest("POST", "/optimize", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response for naming issues
	var res OptimizeResponse
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Expecting naming issues
	if len(res.NamingIssues) == 0 {
		t.Error("Expected naming issues in the response, but found none.")
	}

	// Verify specific issues regarding the invalid scenario names
	if !contains(res.NamingIssues, "Scenario 'Invalid' does not follow naming conventions") {
		t.Error("Expected naming issue for 'Invalid' scenario not found.")
	}

	if !contains(res.NamingIssues, "Scenario '' cannot be empty") {
		t.Error("Expected naming issue for empty scenario name not found.")
	}
}

// Helper function to check if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
