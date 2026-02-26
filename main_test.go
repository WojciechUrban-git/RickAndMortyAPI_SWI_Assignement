package main

import (
	"testing"
)

func TestCountPairs(t *testing.T) {
	// Dummy Data
	mockEpisodes := []Episode{
		{
			Name:       "Pilot",
			Characters: []string{"Rick", "Morty", "Summer"},
		},
		{
			Name:       "Lawnmower Dog",
			Characters: []string{"Rick", "Morty"},
		},
	}

	counts := CountPairs(mockEpisodes)

	if counts["Morty|Rick"] != 2 {
		t.Errorf("Expected Rick and Morty to have 2 episodes, got %d", counts["Morty|Rick"])
	}

	if counts["Rick|Summer"] != 1 {
		t.Errorf("Expected Rick and Summer to have 1 episode, got %d", counts["Rick|Summer"])
	}
}
