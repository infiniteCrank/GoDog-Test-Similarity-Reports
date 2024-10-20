package main

import (
	"fmt"
	"go-similarity-reports/analysis"
	"go-similarity-reports/optimize"
	"go-similarity-reports/visualizations"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Setup routing with Gorilla Mux
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/similarity-reports", analysis.GetSimilarityReports).Methods("GET")
	router.HandleFunc("/api/test-journeys", visualizations.GetTestJourneys).Methods("GET")
	router.HandleFunc("/api/merged-test-journeys", visualizations.GetMergedTestJourneys).Methods("GET")

	router.HandleFunc("/optimize", optimize.OptimizeFeatureHandler).Methods("POST")
	router.HandleFunc("/analyze", analysis.HandleGherkin).Methods("POST")

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
