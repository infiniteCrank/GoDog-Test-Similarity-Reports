package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

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

// Endpoint to get test journeys
func getTestJourneys(w http.ResponseWriter, _ *http.Request) {
	testsPath := "./tdata" // Adjust this path as necessary
	tests, err := parseFeatureFiles(testsPath)
	if err != nil {
		http.Error(w, "Error parsing tests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	testJourneys := generateTestJourneys(tests)

	// Set header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testJourneys)
}

// Endpoint to get similarity reports
func getSimilarityReports(w http.ResponseWriter, r *http.Request) {
	testsPath := "./tdata" // Adjust this path as necessary
	tests, err := parseFeatureFiles(testsPath)
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

// Setup routing with Gorilla Mux
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/similarity-reports", getSimilarityReports).Methods("GET")
	router.HandleFunc("/api/test-journeys", getTestJourneys).Methods("GET")

	// Start the server
	port := "8080"
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
}
