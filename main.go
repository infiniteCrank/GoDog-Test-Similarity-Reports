package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

// Struct for request body to enable feature toggles
type OptimizationOptions struct {
	EnableNamingChecks bool `json:"enableNamingChecks"`
}

// Function to validate scenario names and steps
func validateNamingConvention(scenarios []*Test) error {
	for _, scenario := range scenarios {
		if strings.TrimSpace(scenario.Name) == "" {
			return fmt.Errorf("scenario has an empty name")
		}
		for _, step := range scenario.Steps {
			if len(step) > 80 { // Check for step length limit
				return fmt.Errorf("step exceeds maximum length: %s", step)
			}
		}
	}
	return nil
}

// Function to optimize a feature content
func optimizeFeature(content string, opts OptimizationOptions) string {
	lines := strings.Split(content, "\n")
	var optimizedLines []string
	var scenarios []*Test
	var background []string
	var currentScenario Test

	// Parse scenarios from content
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Scenario:") {
			if currentScenario.Name != "" {
				scenarios = append(scenarios, &currentScenario)
				currentScenario = Test{}
			}
			currentScenario.Name = line
			currentScenario.Steps = []string{}
		} else if strings.HasPrefix(line, "Given") || strings.HasPrefix(line, "When") || strings.HasPrefix(line, "Then") {
			currentScenario.Steps = append(currentScenario.Steps, line)
		} else if strings.HasPrefix(line, "Background:") {
			background = append(background, line)
		} else if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "#") {
			// Handle tags and comments, possibly storing for later use
			optimizedLines = append(optimizedLines, line)
		}
	}

	// Add the last scenario
	if currentScenario.Name != "" {
		scenarios = append(scenarios, &currentScenario)
	}

	// Validate naming conventions if enabled
	if opts.EnableNamingChecks {
		if err := validateNamingConvention(scenarios); err != nil {
			return fmt.Sprintf("Naming convention error: %s", err.Error())
		}
	}

	// Use Background if applicable
	if len(background) > 0 {
		optimizedLines = append(optimizedLines, background...)
	}

	// Consolidate scenarios into example tables
	consolidatedScenarios := map[string]*Test{}
	for _, scenario := range scenarios {
		stepKey := strings.Join(scenario.Steps, "|")
		if existingScenario, found := consolidatedScenarios[stepKey]; found {
			// Combine scenarios into example table
			existingScenario.Name = "Consolidated Scenario with Examples"
		} else {
			consolidatedScenarios[stepKey] = scenario
		}
	}

	// Create an example table
	var exampleTable []string
	for _, scenario := range consolidatedScenarios {
		// Example row format can be customized based on the scenario
		exampleRow := "| param | result |\n"
		exampleRow += "| value1 | result1 |\n"
		exampleRow += "| value2 | result2 |\n"
		exampleTable = append(exampleTable, scenario.Name)
	}

	// Append the example table if we have any consolidated scenarios
	if len(exampleTable) > 0 {
		optimizedLines = append(optimizedLines, "Examples:")
		optimizedLines = append(optimizedLines, strings.Join(exampleTable, "\n"))
	}

	return strings.Join(optimizedLines, "\n")
}

// Endpoint to optimize a feature file with options
func optimizeFeatureFile(w http.ResponseWriter, r *http.Request) {
	var opts OptimizationOptions
	// Decode JSON request body to get optimization options
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Expect a feature file as multipart/form-data
	file, _, err := r.FormFile("featureFile")
	if err != nil {
		http.Error(w, "Error uploading file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the content of the feature file
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Optimize the feature file content
	optimizedContent := optimizeFeature(string(content), opts)

	// Set header and return optimized feature content
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(optimizedContent))
}

type Test struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

type SimilarityReport struct {
	SimilarityType string            `json:"similarity_type"`
	Comparisons    []ComparisonEntry `json:"comparisons"`
}

type ComparisonEntry struct {
	TestA      string  `json:"test_a"`
	TestB      string  `json:"test_b"`
	Similarity float64 `json:"similarity"`
}

type JourneyNode struct {
	Name     string        `json:"name"`
	Children []JourneyNode `json:"children,omitempty"`
}

// Parse feature files in the specified directory
func parseFeatureFiles(path string) ([]Test, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var tests []Test
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".feature") {
			content, _ := os.ReadFile(path + "/" + file.Name())
			re := regexp.MustCompile(`(?m)^\s*(Given|When|Then)\s+(.*)`)
			matches := re.FindAllStringSubmatch(string(content), -1)

			var steps []string
			for _, match := range matches {
				steps = append(steps, match[2])
			}
			tests = append(tests, Test{Name: file.Name(), Steps: steps})
		}
	}
	return tests, nil
}

// Calculate Longest Common Subsequence (LCS)
func LCS(X, Y []string) int {
	m := len(X)
	n := len(Y)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if X[i-1] == Y[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp[m][n]
}

// Calculate Cosine Similarity
func CosineSimilarity(testA, testB []string) float64 {
	stepCount := make(map[string]int)

	for _, step := range testA {
		stepCount[step]++
	}
	for _, step := range testB {
		stepCount[step]++
	}

	var vectorA, vectorB []float64
	for _, count := range stepCount {
		vectorA = append(vectorA, float64(count))
		if count > 1 {
			vectorB = append(vectorB, float64(count))
		}
	}

	if len(vectorA) == 0 || len(vectorB) == 0 {
		return 0.0
	}

	dotProduct := 0.0
	magA := 0.0
	magB := 0.0
	for i := range vectorA {
		dotProduct += vectorA[i] * vectorB[i]
		magA += vectorA[i] * vectorA[i]
		magB += vectorB[i] * vectorB[i]
	}
	return dotProduct / (math.Sqrt(magA) * math.Sqrt(magB))
}

// Calculate Jaccard Index
func JaccardIndex(setA, setB []string) float64 {
	setASet := make(map[string]struct{})
	setBSet := make(map[string]struct{})
	for _, step := range setA {
		setASet[step] = struct{}{}
	}
	for _, step := range setB {
		setBSet[step] = struct{}{}
	}

	intersectionSize := 0
	for step := range setASet {
		if _, found := setBSet[step]; found {
			intersectionSize++
		}
	}

	unionSize := len(setASet) + len(setBSet) - intersectionSize
	return float64(intersectionSize) / float64(unionSize)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Generate test journey hierarchy
func generateTestJourneys(tests []Test) JourneyNode {
	root := JourneyNode{Name: "Test Journeys", Children: []JourneyNode{}}
	for _, test := range tests {
		testNode := JourneyNode{Name: test.Name, Children: []JourneyNode{}}
		for _, step := range test.Steps {
			stepNode := JourneyNode{Name: step}
			testNode.Children = append(testNode.Children, stepNode)
		}
		root.Children = append(root.Children, testNode)
	}
	return root
}

// Endpoint to get similarity reports
func getSimilarityReports(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("directory")
	if dir == "" {
		dir = "./tdata" // Default path
	}

	tests, err := parseFeatureFiles(dir)
	if err != nil {
		http.Error(w, "Error parsing tests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lcsReport := SimilarityReport{SimilarityType: "LCS", Comparisons: []ComparisonEntry{}}
	cosineReport := SimilarityReport{SimilarityType: "Cosine Similarity", Comparisons: []ComparisonEntry{}}
	jaccardReport := SimilarityReport{SimilarityType: "Jaccard Index", Comparisons: []ComparisonEntry{}}

	for i := 0; i < len(tests); i++ {
		for j := i + 1; j < len(tests); j++ {
			// Calculate LCS
			lcs := LCS(tests[i].Steps, tests[j].Steps)
			lcsSimilarity := float64(lcs) / float64(len(tests[i].Steps)+len(tests[j].Steps)-lcs)
			lcsReport.Comparisons = append(lcsReport.Comparisons, ComparisonEntry{
				TestA:      tests[i].Name,
				TestB:      tests[j].Name,
				Similarity: lcsSimilarity,
			})

			// Calculate Cosine Similarity
			cosineSimilarity := CosineSimilarity(tests[i].Steps, tests[j].Steps)
			cosineReport.Comparisons = append(cosineReport.Comparisons, ComparisonEntry{
				TestA:      tests[i].Name,
				TestB:      tests[j].Name,
				Similarity: cosineSimilarity,
			})

			// Calculate Jaccard Index
			jaccardSimilarity := JaccardIndex(tests[i].Steps, tests[j].Steps)
			jaccardReport.Comparisons = append(jaccardReport.Comparisons, ComparisonEntry{
				TestA:      tests[i].Name,
				TestB:      tests[j].Name,
				Similarity: jaccardSimilarity,
			})
		}
	}

	// Prepare the response
	response := struct {
		LCSReport     SimilarityReport `json:"lcs_report"`
		CosineReport  SimilarityReport `json:"cosine_report"`
		JaccardReport SimilarityReport `json:"jaccard_report"`
	}{
		LCSReport:     lcsReport,
		CosineReport:  cosineReport,
		JaccardReport: jaccardReport,
	}

	// Set header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Endpoint to get test journeys
func getTestJourneys(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("directory")
	if dir == "" {
		dir = "./tdata" // Default path
	}

	tests, err := parseFeatureFiles(dir)
	if err != nil {
		http.Error(w, "Error parsing tests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	testJourneys := generateTestJourneys(tests)

	// Set header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testJourneys)
}

// Setup routing with Gorilla Mux
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/similarity-reports", getSimilarityReports).Methods("GET")
	router.HandleFunc("/api/test-journeys", getTestJourneys).Methods("GET")
	router.HandleFunc("/api/optimize-feature", optimizeFeatureFile).Methods("POST")

	// Serve static files from the public directory
	fs := http.FileServer(http.Dir("public"))
	router.PathPrefix("/").Handler(fs)

	// Start the server
	port := "8080"
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
