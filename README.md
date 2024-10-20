# GoDog-Test-Similarity-Reports
This GoLang App takes in a given directory and scans GoDog files and tries to find test that are similar to reduce test cases in large test suites 

## File Structure 
go-similarity-reports/ \
├── go.mod \
├── go.sum \
├── main.go \
└── tdata/ \
............└── # Your .feature files go here


## Run the Program:
From the terminal in the go-similarity-reports directory, run: \
```go run main.go```

## Access the Similarity Reports:
Once the server is running, you can access the similarity reports by navigating to:
http://localhost:8080/api/similarity-reports?directory=./your-directory


## Example of a JSON Response
When you access the /api/similarity-reports endpoint, you should receive a response similar to the following (assuming there are feature files with steps):

```{
  "lcs_report": {
    "similarity_type": "LCS",
    "comparisons": [
      {
        "test_a": "test1.feature",
        "test_b": "test2.feature",
        "similarity": 0.75
      },
      ...
    ]
  },
  "cosine_report": {
    "similarity_type": "Cosine Similarity",
    "comparisons": [
      {
        "test_a": "test1.feature",
        "test_b": "test2.feature",
        "similarity": 0.82
      },
      ...
    ]
  },
  "jaccard_report": {
    "similarity_type": "Jaccard Index",
    "comparisons": [
      {
        "test_a": "test1.feature",
        "test_b": "test2.feature",
        "similarity": 0.5
      },
      ...
    ]
  }
}
```
## Explanation of Similarity report  

### Cosine Similarity Report (cosine_report)
Method: Cosine similarity measures the cosine of the angle between two non-zero vectors in a multi-dimensional space. It is defined as the dot product of the vectors divided by the product of their magnitudes.

 - Use Case: Cosine similarity is particularly effective for high-dimensional data where the direction of the data matters more than the magnitude. In this context, it helps compare the frequency of occurrence of each step in the test cases.

 - Range: The result ranges between -1 and 1, where:

 - 1 means the vectors are identical,
0 means the vectors are orthogonal (not similar),
-1 means the vectors are diametrically opposed (completely dissimilar).
Example: Two tests that share many common steps will have a high cosine similarity score, while those with very few shared steps will have a lower score.

### Jaccard Index Report (jaccard_report)
Method: The Jaccard index measures similarity by comparing the size of the intersection of two sets to the size of their union. It is defined as the size of the intersection divided by the size of the union of the sample sets.

 - Use Case: The Jaccard index is suitable for binary data or situations where the presence or absence of elements matters (like steps being present or not). It is often used in scenarios like clustering, finding duplicates, and comparing binary attributes.

 - Range: The Jaccard index ranges from 0 to 1:

 - 1 means the two sets are identical (i.e., they have the same elements),
0 means there are no common elements at all.
Example: If one test has steps {A, B, C} and another has {B, C, D}, the Jaccard index would quantify how similar these two sets of steps are.

### Longest Common Subsequence Report (lcs_report)
Method: The Longest Common Subsequence (LCS) is a classic dynamic programming technique. It finds the longest subsequence present in both sequences. The LCS does not require the substrings to be contiguous, only in the same order.

 - Use Case: LCS is useful for analyzing sequences where the order of elements is crucial. It can be used in applications like version control, DNA sequence analysis, and in evaluating similar text documents.

 - Range: The LCS score itself is the length of the common subsequence, which is then often normalized to the length of both input sequences (similar to cosine or Jaccard) to generate a similarity score:

 - The score is typically presented as a ratio reflecting the similarity between the sequences.
Example: Given two tests where steps are {A, B, C} and {B, C, D}, the LCS would be "BC," and the similarity can be expressed based on the length of the LCS relative to the lengths of both test sequences.

## Summary
 - Cosine Similarity is best for frequency-based metrics and high-dimensional spaces.

 - Jaccard Index is ideal for binary presence/absence analysis of sets.

 - Longest Common Subsequence is apt for comparing ordered sequences and is particularly effective when the order of elements matters.

These three metrics provide diverse perspectives on similarity, allowing you to assess the duplicates or related tests according to different characteristics of their steps. Depending on your testing strategy and the types of inputs, you might choose one or more of these methods to determine test case similarities.


## Test Journey Hierarchy Endpoint
http://localhost:8080/api/test-journeys?directory=./your-directory

Produces a hierarchy Data Structure: A new JourneyNode struct was created to represent each test and its steps in a hierarchical structure for use with D3.js.

# Gherkin Feature File Optimizer

## Overview

This application optimizes Gherkin feature files by merging scenarios with common steps, creating a `Background` section, and checking scenario naming conventions. It also preserves tags found in the original feature file.

## Features

- **Background Generation**: Generates a `Background` section for common steps shared by scenarios.
- **Tag Preservation**: Maintains tags in the optimized output.
- **Naming Convention Check**: Validates that scenario names conform to recommended practices.
- **Togglable Functionality**: Users can toggle the naming convention checks.

## Requirements

- Go (version 1.14 or higher)
- Suitable gherkin parser (included in the project)


## Force-directed graph Explanation  

### Node and Link Creation: 

 - The graph is built using nodes representing test cases and links showing the similarities between them. Nodes and links are created based on the similarity reports (LCS, Cosine, Jaccard).
 - The strength of each link is derived from the similarity score, influencing its stroke width.

### Simulation: 

 - A D3 force simulation is created that handles the physics of the graph. The nodes are repelled from each other, and links pull connected nodes together, resulting in an organized layout:
   - Force Link: Keeps nodes connected.
   - Charge: Adjusts the spacing between nodes.
   - Center: Centers the graph in the SVG area.

### Tick Function: 

- The ```ticked``` function updates the positions of nodes and links on each tick of the simulation, ensuring they move fluidly as the forces act upon them.

### Dragging Functionality:

 - Allows users to click and drag nodes around in the graph, powered by the D3 drag functionality.