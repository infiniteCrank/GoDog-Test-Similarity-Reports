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

// HandleGherkin processes incoming HTTP requests to analyze Gherkin files.
func HandleGherkin(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to analyze Gherkin file.") // Log the receipt of a request.

	// Read the Gherkin file content from the request body.
	body, err := io.ReadAll(r.Body) // Read all data from the request body into a byte slice.
	if err != nil {                 // Check for errors during reading.
		http.Error(w, "Failed reading body", http.StatusBadRequest) // Respond with a 400 Bad Request error.
		log.Println("Error reading request body:", err)             // Log the error for troubleshooting.
		return                                                      // Exit the function to prevent further processing.
	}
	defer r.Body.Close() // Ensure the request body is closed after reading to free up resources.

	reader := strings.NewReader(string(body)) // Create a new strings.Reader to facilitate reading the body as a string.
	uuid := &messages.UUID{}                  // Initialize a new UUID for unique identification.

	// Analyze the Gherkin document for scenarios and their probabilities.
	results, err := analyzeGherkinDocument(reader, uuid)
	if err != nil { // Check for any parsing errors.
		http.Error(w, "Failed to parse Gherkin document", http.StatusInternalServerError) // Respond with a 500 Internal Server Error.
		log.Println("Error parsing Gherkin document:", err)                               // Log the error for debugging purposes.
		return                                                                            // Exit the function to prevent further processing.
	}

	w.Header().Set("Content-Type", "application/json")         // Set the response content type to JSON.
	if err := json.NewEncoder(w).Encode(results); err != nil { // Encode the results into JSON and write to the response.
		http.Error(w, "Failed to encode response", http.StatusInternalServerError) // Respond with a 500 Internal Server Error.
		log.Println("Error encoding JSON response:", err)                          // Log any encoding errors.
	}

	log.Println("Successfully analyzed Gherkin file and returned results.") // Log the successful processing of the request.
}
