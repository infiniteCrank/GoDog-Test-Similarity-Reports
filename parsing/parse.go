package parsing

import (
	"os"
	"regexp"
	"strings"
)

type Test struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

// Parse feature files in the specified directory
func ParseFeatureFiles(path string) ([]Test, error) {
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
