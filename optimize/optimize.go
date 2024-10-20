package optimize

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
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
		if len(scenario.Name) < 10 { // Example check for length
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

		// Check if the scenario already exists in the map
		if existingScenario, found := scenarioMap[key]; found {
			// If it exists, we can append data but do not overwrite existing steps
			existingScenario.Data = append(existingScenario.Data, scenario.Data...)
			existingScenario.Tags = append(existingScenario.Tags, scenario.Tags...) // Merge tags
		} else {
			newScenario := &Scenario{
				Name:  scenario.Name,
				Steps: scenario.Steps,
				Tags:  scenario.Tags,
				Data:  scenario.Data, // Initialize with scenario Data
			}
			scenarioMap[key] = newScenario
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

// OptimizeFeatureHandler handles the upload and optimization of one or more feature files
func OptimizeFeatureHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // Limit your max input length!
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	var allScenarios []Scenario
	errorsChan := make(chan error, 10) // Channel for error handling
	var wg sync.WaitGroup              // WaitGroup to manage goroutines

	// Process multiple uploaded files
	for _, fheaders := range r.MultipartForm.File {
		for _, file := range fheaders {
			wg.Add(1) // Increment wait group counter

			go func(fh *multipart.FileHeader) {
				defer wg.Done() // Decrement counter when done

				// Open the uploaded file
				uploadedFile, err := fh.Open()
				if err != nil {
					errorsChan <- fmt.Errorf("error opening file: %s, %v", fh.Filename, err)
					return
				}
				defer uploadedFile.Close()

				// Read the content of the file
				content, err := io.ReadAll(uploadedFile)
				if err != nil {
					errorsChan <- fmt.Errorf("error reading file content: %s, %v", fh.Filename, err)
					return
				}

				// Split content into lines to parse scenarios
				lines := strings.Split(string(content), "\n")
				var currentScenario Scenario

				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "Scenario:") {
						// Finish collecting the previous scenario
						if currentScenario.Name != "" {
							allScenarios = append(allScenarios, currentScenario) // Save previous scenario
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
					allScenarios = append(allScenarios, currentScenario)
				}
			}(file) // Pass the file header to the goroutine
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errorsChan) // Close the channel when done

	// Check for errors after processing all files
	for err := range errorsChan {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate the feature name
	featureName := "Combined Features" // Set to determine the relevant feature title
	if err := validateFeatureName(featureName); err != nil {
		http.Error(w, fmt.Sprintf("Feature validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Check naming conventions if enabled
	checkNaming := r.FormValue("check_naming") == "true"
	var namingIssues []string
	if checkNaming {
		namingIssues = validateScenarioNames(allScenarios)
	}

	// Optimize scenarios and prepare optimized content
	optimizedScenarios, commonSteps := optimizeScenarios(allScenarios)
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
