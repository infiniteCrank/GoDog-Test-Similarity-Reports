package analysis

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v27"
	messages "github.com/cucumber/messages/go/v22"
	"github.com/jdkato/prose/v2"
)

func analyzeScenarioText(scenarioText string) map[string]float64 {
	doc, err := prose.NewDocument(scenarioText)
	if err != nil {
		return nil // Handle error (consider logging)
	}

	keywords := map[string]int{}
	for _, tok := range doc.Tokens() {
		if tok.Tag == "NN" { // Nouns
			keywords[tok.Text]++
		} else if tok.Tag == "VB" { // Verbs
			keywords[tok.Text]++
		}
	}

	// Calculate probabilities based on keywords found
	probabilities := map[string]float64{
		"component":   determineComponentProbability(keywords),
		"integration": determineIntegrationProbability(keywords),
		"end_to_end":  determineEndToEndProbability(keywords),
		"regression":  determineRegressionProbability(keywords),
	}
	return probabilities
}

// Example probability calculations (customize as needed)
func determineComponentProbability(keywords map[string]int) float64 {
	// Implement your logic based on keywords, example:
	if keywords["button"] > 0 {
		return 0.8
	}
	return 0.2
}

func determineIntegrationProbability(keywords map[string]int) float64 {
	if keywords["API"] > 0 || keywords["connect"] > 0 {
		return 0.7
	}
	return 0.3
}

func determineEndToEndProbability(keywords map[string]int) float64 {
	if keywords["user"] > 0 && keywords["login"] > 0 {
		return 0.6
	}
	return 0.2
}

func determineRegressionProbability(keywords map[string]int) float64 {
	if keywords["previous"] > 0 {
		return 0.5
	}
	return 0.1
}

// Structure for output
type ScenarioProbability struct {
	ScenarioName string             `json:"scenario_name"`
	Probability  map[string]float64 `json:"probability"`
}

func analyzeGherkinDocument(reader *strings.Reader, uuid *messages.UUID) ([]ScenarioProbability, error) {
	gherkinDocument, err := gherkin.ParseGherkinDocument(reader, uuid.NewId)
	if err != nil {
		return nil, err
	}

	pickles := gherkin.Pickles(*gherkinDocument, "minimal.feature", uuid.NewId)

	var results []ScenarioProbability
	for _, pickle := range pickles {
		// Analyze the scenario text for probabilities
		probabilities := analyzeScenarioText(pickle.Name)
		results = append(results, ScenarioProbability{
			ScenarioName: pickle.Name,
			Probability:  probabilities,
		})
	}

	return results, nil
}

func HandleGherkin(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to analyze Gherkin file.")

	// Read the Gherkin file content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed reading body", http.StatusBadRequest)
		log.Println("Error reading request body:", err)
		return
	}
	defer r.Body.Close()

	reader := strings.NewReader(string(body))
	uuid := &messages.UUID{}

	results, err := analyzeGherkinDocument(reader, uuid)
	if err != nil {
		http.Error(w, "Failed to parse Gherkin document", http.StatusInternalServerError)
		log.Println("Error parsing Gherkin document:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Println("Error encoding JSON response:", err)
	}
	log.Println("Successfully analyzed Gherkin file and returned results.")
}
