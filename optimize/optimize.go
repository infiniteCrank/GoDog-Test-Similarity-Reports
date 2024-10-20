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
	Tags  []string
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

// Optimize scenarios, including support for Scenario Outlines with Examples
func optimizeScenarios(scenarios []Scenario) ([]Scenario, []string) {
	scenarioMap := make(map[string]*Scenario)

	for _, scenario := range scenarios {
		// If it's a Scenario Outline, handle the examples
		if len(scenario.Data) > 0 {
			for _, example := range scenario.Data {
				newScenario := Scenario{
					Name:  scenario.Name,
					Steps: scenario.Steps,
					Tags:  scenario.Tags,
				}

				// Replace placeholders in steps with actual example values
				for i, step := range newScenario.Steps {
					for key, value := range example {
						step = strings.ReplaceAll(step, key, value)
					}
					newScenario.Steps[i] = step
				}

				key := strings.Join(newScenario.Steps, "|")
				if existingScenario, found := scenarioMap[key]; found {
					// Merge scenario data if it already exists
					existingScenario.Data = append(existingScenario.Data, example)
				} else {
					newScenario.Data = []map[string]string{example}
					scenarioMap[key] = &newScenario
				}
			}
		} else {
			// Regular scenario handling
			key := strings.Join(scenario.Steps, "|")
			if existingScenario, found := scenarioMap[key]; found {
				existingScenario.Data = append(existingScenario.Data, nil) // May contain example data
			} else {
				scenarioMap[key] = &scenario // No examples
			}
		}
	}

	optimizedScenarios := make([]Scenario, 0, len(scenarioMap))
	for _, v := range scenarioMap {
		optimizedScenarios = append(optimizedScenarios, *v)
	}

	// Identify common steps for Background
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

	// Iterate over the optimized scenarios and write them
	for _, scenario := range optimizedScenarios {
		output.WriteString("Scenario: " + scenario.Name + "\n")

		// Write steps for each scenario
		for _, step := range scenario.Steps {
			output.WriteString("  " + step + "\n")
		}
	}

	return output.String()
}

// OptimizeFeatureHandler handles the upload and optimization of the feature files
func OptimizeFeatureHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // Limit your max input length!
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form input
	file, _, err := r.FormFile("feature_file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the content of the file
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
			// Finish up the previous scenario
			if currentScenario.Name != "" {
				scenarios = append(scenarios, currentScenario) // Save previous scenario
			}
			currentScenario = Scenario{Name: strings.TrimSpace(strings.TrimPrefix(trimmed, "Scenario:"))}
			currentScenario.Steps = []string{}
		} else if strings.HasPrefix(trimmed, "@") {
			// Handle tags
			currentScenario.Tags = append(currentScenario.Tags, strings.TrimSpace(trimmed))
		} else if strings.HasPrefix(trimmed, "Given") ||
			strings.HasPrefix(trimmed, "When") ||
			strings.HasPrefix(trimmed, "Then") {
			currentScenario.Steps = append(currentScenario.Steps, trimmed)
		}
	}

	// Append the last collected scenario if it exists
	if currentScenario.Name != "" {
		scenarios = append(scenarios, currentScenario)
	}

	// Validate the feature name (this could be more dynamic)
	featureName := "User Login" // Set it to determine the relevant feature title
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

	// Prepare response
	response := OptimizeResponse{
		OptimizedContent: optimizedContent,
		NamingIssues:     namingIssues,
	}

	// Send the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
