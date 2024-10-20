package visualizations

import (
	"encoding/json"
	"go-similarity-reports/parsing"
	"net/http"
)

type JourneyNode struct {
	Name     string        `json:"name"`
	Children []JourneyNode `json:"children,omitempty"`
}

// Function to merge identical nodes in the test journeys
func mergeIdenticalNodes(nodes []parsing.Test) []*parsing.Test {
	nodeMap := make(map[string]*parsing.Test) // Map to hold unique nodes

	for _, node := range nodes {
		if existingNode, found := nodeMap[node.Name]; found {
			// If the node name already exists, merge the steps
			existingNode.Steps = append(existingNode.Steps, node.Steps...)
		} else {
			// Otherwise, save the new node
			nodeMap[node.Name] = &parsing.Test{
				Name:  node.Name,
				Steps: node.Steps,
			}
		}
	}

	// Convert map values to slice
	mergedNodes := []*parsing.Test{}
	for _, node := range nodeMap {
		mergedNodes = append(mergedNodes, node)
	}

	return mergedNodes
}

// New endpoint to get test journeys with merged identical nodes
func GetMergedTestJourneys(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("directory")
	if dir == "" {
		dir = "./tdata" // Default path
	}

	tests, err := parsing.ParseFeatureFiles(dir)
	if err != nil {
		http.Error(w, "Error parsing tests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Merge identical nodes across scenarios
	mergedTests := mergeIdenticalNodes(tests)

	// Prepare response with the merged test journeys
	response := struct {
		Name     string          `json:"name"`
		Children []*parsing.Test `json:"children"`
	}{
		Name:     "Merged Test Journeys",
		Children: mergedTests,
	}

	// Set header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Generate test journey hierarchy
func generateTestJourneys(tests []parsing.Test) JourneyNode {
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

// Endpoint to get test journeys
func GetTestJourneys(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("directory")
	if dir == "" {
		dir = "./tdata" // Default path
	}

	tests, err := parsing.ParseFeatureFiles(dir)
	if err != nil {
		http.Error(w, "Error parsing tests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	testJourneys := generateTestJourneys(tests)

	// Set header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testJourneys)
}
