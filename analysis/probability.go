package analysis

import (
	"encoding/json" // Provides functions for encoding and decoding JSON.
	"io"            // Provides functions for I/O operations.
	"log"           // Provides logging functions.
	"net/http"      // Provides HTTP client and server implementations.
	"strings"       // Provides string manipulation functions.

	gherkin "github.com/cucumber/gherkin/go/v27"   // Go library for parsing Gherkin files.
	messages "github.com/cucumber/messages/go/v22" // Go library for handling cucumber messages.
	"github.com/jdkato/prose/v2"                   // NLP library for processing text.
)

// analyzeScenarioText analyzes the text of a Gherkin scenario and returns the estimated probabilities for each test type.
func analyzeScenarioText(scenarioText string) map[string]float64 {
	doc, err := prose.NewDocument(scenarioText) // Create a new document for NLP processing.
	if err != nil {
		return nil // Return nil if there's an error creating the document (consider logging this).
	}

	keywords := map[string]int{}       // Initialize a map to count the occurrences of keywords.
	for _, tok := range doc.Tokens() { // Iterate over each token in the document.
		if tok.Tag == "NN" { // Check if the token is a noun.
			keywords[tok.Text]++ // Increment the count for the noun.
		} else if tok.Tag == "VB" { // Check if the token is a verb.
			keywords[tok.Text]++ // Increment the count for the verb.
		}
	}

	// Calculate probabilities for test types based on the keywords found.
	probabilities := map[string]float64{
		"component":   determineComponentProbability(keywords),   // Calculate component test probability.
		"integration": determineIntegrationProbability(keywords), // Calculate integration test probability.
		"end_to_end":  determineEndToEndProbability(keywords),    // Calculate end-to-end test probability.
		"regression":  determineRegressionProbability(keywords),  // Calculate regression test probability.
	}
	return probabilities // Return the calculated probabilities.
}

// Example probability calculations for component tests based on keywords.
func determineComponentProbability(keywords map[string]int) float64 {
	// If the keyword "button" appears, assume it's highly likely to be a component test.
	if keywords["button"] > 0 {
		return 0.8
	}
	return 0.2 // Default low probability if "button" is not found.
}

// Example probability calculations for integration tests based on keywords.
func determineIntegrationProbability(keywords map[string]int) float64 {
	// If "API" or "connect" keywords appear, assume a high likelihood of integration tests.
	if keywords["API"] > 0 || keywords["connect"] > 0 {
		return 0.7
	}
	return 0.3 // Default low probability otherwise.
}

// Example probability calculations for end-to-end tests based on keywords.
func determineEndToEndProbability(keywords map[string]int) float64 {
	// If both "user" and "login" keywords appear, assume a moderate likelihood for end-to-end tests.
	if keywords["user"] > 0 && keywords["login"] > 0 {
		return 0.6
	}
	return 0.2 // Default low probability otherwise.
}

// Example probability calculations for regression tests based on keywords.
func determineRegressionProbability(keywords map[string]int) float64 {
	// If the keyword "previous" is found, assume moderate likelihood for regression tests.
	if keywords["previous"] > 0 {
		return 0.5
	}
	return 0.1 // Default low probability otherwise.
}

// Structure for output representing a scenario and its associated probabilities.
type ScenarioProbability struct {
	ScenarioName string             `json:"scenario_name"` // Name of the scenario.
	Probability  map[string]float64 `json:"probability"`   // Probabilities for each test type.
}

// analyzeGherkinDocument parses a Gherkin document and analyzes its scenarios for test type probabilities.
func analyzeGherkinDocument(reader *strings.Reader, uuid *messages.UUID) ([]ScenarioProbability, error) {
	// Parse the Gherkin document from the reader.
	gherkinDocument, err := gherkin.ParseGherkinDocument(reader, uuid.NewId)
	if err != nil {
		return nil, err // Return nil and the error if parsing fails.
	}

	// Get the pickles (individual scenarios) from the Gherkin document.
	pickles := gherkin.Pickles(*gherkinDocument, "minimal.feature", uuid.NewId)

	var results []ScenarioProbability // Initialize a slice to hold the results.
	for _, pickle := range pickles {  // Iterate over each extracted scenario (pickle).
		// Analyze the scenario text for probabilities.
		probabilities := analyzeScenarioText(pickle.Name)
		// Append
		results = append(results, ScenarioProbability{ // Create a ScenarioProbability object.
			ScenarioName: pickle.Name,   // Set the scenario name.
			Probability:  probabilities, // Set the calculated probabilities.
		})
	}

	return results, nil // Return the aggregated results and nil error.
}

func HandleGherkin(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to analyze Gherkin files.")

	// Parse the multipart form, with a max memory of 10 MB
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		log.Println("Error parsing multipart form:", err)
		return
	}

	files := r.MultipartForm.File["files"] // Collect the uploaded files from the form

	var results []ScenarioProbability // Initialize a slice to hold results for each file

	for _, f := range files {
		// Open the uploaded file
		file, err := f.Open()
		if err != nil {
			http.Error(w, "Failed to open uploaded file", http.StatusInternalServerError)
			log.Println("Error opening file:", err)
			return
		}
		defer file.Close() // Ensure the file is closed after processing

		// Read the file contents
		body, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file contents", http.StatusInternalServerError)
			log.Println("Error reading file contents:", err)
			return
		}

		// Create a new strings.Reader from the file contents
		reader := strings.NewReader(string(body))
		uuid := &messages.UUID{}

		// Analyze the Gherkin document for scenarios and their probabilities
		result, err := analyzeGherkinDocument(reader, uuid)
		if err != nil {
			http.Error(w, "Failed to parse Gherkin document", http.StatusInternalServerError)
			log.Println("Error parsing Gherkin document:", err)
			return
		}

		// Append the results from this file to the main results
		results = append(results, result...)
	}

	// Set the response content type to JSON and encode the results
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Println("Error encoding JSON response:", err)
		return
	}

	log.Println("Successfully analyzed Gherkin files and returned results.")
}
