package optimize

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Scenario struct {
	Name  string
	Steps []string
	Tags  []string            // Include tags for each scenario
	Data  []map[string]string // Holds example data for the scenarios
}

type OptimizeResponse struct {
	OptimizedContent string   `json:"optimized_content"`
	NamingIssues     []string `json:"naming_issues,omitempty"` // Omitempty to avoid sending null
}

// Validate the feature name structure
func validateFeatureName(name string) error {
	if name == "" {
		return errors.New("feature name is required")
	}
	return nil
}

// Validate scenario names against conventions
func validateScenarioNames(scenarios []Scenario) []string {
	var issues []string
	for _, scenario := range scenarios {
		if len(scenario.Name) == 0 {
			issues = append(issues, "Scenario name cannot be empty")
		}
		if len(scenario.Name) < 10 { // Example check; adjust as per your naming conventions
			issues = append(issues, fmt.Sprintf("Scenario '%s' does not follow naming conventions", scenario.Name))
		}
	}
	return issues
}

// Identify common steps for Background
func identifyCommonSteps(scenarios []Scenario) []string {
	if len(scenarios) == 0 {
		return nil
	}

	commonSteps := make(map[string]int)
	for _, scenario := range scenarios {
		for _, step := range scenario.Steps {
			commonSteps[step]++
		}
	}

	var sharedSteps []string
	for step, count := range commonSteps {
		if count == len(scenarios) {
			sharedSteps = append(sharedSteps, step)
		}
	}

	return sharedSteps
}

// Optimize scenarios by merging identical ones
func optimizeScenarios(scenarios []Scenario) ([]Scenario, []string) {
	scenarioMap := make(map[string]*Scenario)

	for _, scenario := range scenarios {
		key := strings.Join(scenario.Steps, "|")

		if existingScenario, found := scenarioMap[key]; found {
			// Append existing only scenario's steps if duplicates are found
			existingScenario.Steps = append(existingScenario.Steps, scenario.Steps...)
		} else {
			newScenario := &Scenario{
				Name:  scenario.Name,
				Steps: scenario.Steps,
				Tags:  scenario.Tags,
			}
			scenarioMap[key] = newScenario
		}
	}

	optimizedScenarios := make([]Scenario, 0, len(scenarioMap))
	for _, v := range scenarioMap {
		optimizedScenarios = append(optimizedScenarios, *v)
	}

	// Identify common steps that can be used in a Background section.
	commonSteps := identifyCommonSteps(scenarios)

	return optimizedScenarios, commonSteps
}

// Generate the optimized content, including Background if applicable
func writeOptimizedContent(featureName string, optimizedScenarios []Scenario, commonSteps []string) string {
	var output bytes.Buffer
	output.WriteString("Feature: " + featureName + "\n")

	// If there are common steps, add Background
	if len(commonSteps) > 0 {
		output.WriteString("Background:\n")
		for _, step := range commonSteps {
			output.WriteString("  " + step + "\n")
		}
	}

	for _, scenario := range optimizedScenarios {
		output.WriteString("Scenario: " + scenario.Name + "\n")

		// Write steps for each scenario
		for _, step := range scenario.Steps {
			output.WriteString("  " + step + "\n")
		}
	}

	return output.String()
}

func OptimizeFeatureHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // Limit your max input length!
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("feature_file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file content", http.StatusInternalServerError)
		return
	}

	// Split content into lines to parse scenarios
	lines := strings.Split(string(content), "\n")
	var scenarios []Scenario
	var currentScenario Scenario

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Scenario:") {
			if currentScenario.Name != "" {
				scenarios = append(scenarios, currentScenario) // Save previous scenario
			}
			currentScenario = Scenario{Name: strings.TrimSpace(strings.TrimPrefix(trimmed, "Scenario:"))}
			currentScenario.Steps = []string{}
		} else if strings.HasPrefix(trimmed, "@") {
			// Handle scenario tags
			currentScenario.Tags = append(currentScenario.Tags, trimmed)
		} else if strings.HasPrefix(trimmed, "Given") ||
			strings.HasPrefix(trimmed, "When") ||
			strings.HasPrefix(trimmed, "Then") {
			currentScenario.Steps = append(currentScenario.Steps, trimmed)
		}
	}

	// Add the last collected scenario
	if currentScenario.Name != "" {
		scenarios = append(scenarios, currentScenario)
	}

	// Validate the feature name (static in this example; you may want to make this dynamic)
	featureName := "User Login" // Static feature name; set appropriately
	if err := validateFeatureName(featureName); err != nil {
		http.Error(w, fmt.Sprintf("Feature validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Check naming conventions if enabled
	checkNaming := r.FormValue("check_naming") == "true"
	var namingIssues []string
	if checkNaming {
		namingIssues = validateScenarioNames(scenarios)
	}

	// Optimize scenarios and prepare optimized content
	optimizedScenarios, commonSteps := optimizeScenarios(scenarios)
	optimizedContent := writeOptimizedContent(featureName, optimizedScenarios, commonSteps)

	// Prepare response structure
	response := OptimizeResponse{
		OptimizedContent: optimizedContent,
		NamingIssues:     namingIssues,
	}

	// Send the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
