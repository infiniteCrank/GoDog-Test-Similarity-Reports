package statistics

import (
	"encoding/json"
	"go-similarity-reports/parsing"
	"math"
	"net/http"
)

type SimilarityReport struct {
	SimilarityType string            `json:"similarity_type"`
	Comparisons    []ComparisonEntry `json:"comparisons"`
}

type ComparisonEntry struct {
	TestA      string  `json:"test_a"`
	TestB      string  `json:"test_b"`
	Similarity float64 `json:"similarity"`
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

func CosineSimilarity(testA, testB []string) float64 {
	// Create frequency maps for steps in both tests
	stepCountA := make(map[string]int)
	stepCountB := make(map[string]int)

	for _, step := range testA {
		stepCountA[step]++
	}
	for _, step := range testB {
		stepCountB[step]++
	}

	// Create unique combined set of steps
	uniqueSteps := make(map[string]bool)
	for step := range stepCountA {
		uniqueSteps[step] = true
	}
	for step := range stepCountB {
		uniqueSteps[step] = true
	}

	// Create vectors based on unique steps
	var vectorA, vectorB []float64
	for step := range uniqueSteps {
		vectorA = append(vectorA, float64(stepCountA[step])) // value for test A
		vectorB = append(vectorB, float64(stepCountB[step])) // value for test B
	}

	// Compute cosine similarity
	dotProduct := 0.0
	magA := 0.0
	magB := 0.0
	for i := 0; i < len(vectorA); i++ {
		dotProduct += vectorA[i] * vectorB[i] // dot product
		magA += vectorA[i] * vectorA[i]       // magnitude of A
		magB += vectorB[i] * vectorB[i]       // magnitude of B
	}

	// Handle cases where the magnitude is 0
	if magA == 0 || magB == 0 {
		return 0.0 // If either vector has no steps, similarity is undefined, return 0
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

// Endpoint to get similarity reports
func GetSimilarityReports(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("directory")
	if dir == "" {
		dir = "./tdata" // Default path
	}

	tests, err := parsing.ParseFeatureFiles(dir)
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
