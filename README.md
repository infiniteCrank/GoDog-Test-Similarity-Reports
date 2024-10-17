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

## Access the API:
Once the server is running, you can access the similarity reports by navigating to:
http://localhost:8080/api/similarity-reports

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