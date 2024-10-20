package analysis

import (
	"testing"
)

func TestDetermineComponentProbability(t *testing.T) {
	keywords := map[string]int{"button": 1}
	expected := 0.8
	result := determineComponentProbability(keywords)
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestDetermineIntegrationProbability(t *testing.T) {
	keywords := map[string]int{"API": 1}
	expected := 0.7
	result := determineIntegrationProbability(keywords)
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

// Add tests for other functions similarly
