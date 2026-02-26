package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// SearchResult is what we send back to the frontend.
type SearchResult struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// RawResponse helps us parse the results array from the external API.
type RawResponse struct {
	Results []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type Episode struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Characters []string `json:"characters"`
}

type CharacterInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// PairResult represents two characters and how many times they shared a scene.
type PairResult struct {
	Character1 CharacterInfo `json:"character1"`
	Character2 CharacterInfo `json:"character2"`
	Episodes   int           `json:"episodes"`
}

type EpisodeResponse struct {
	Info struct {
		Next string `json:"next"`
	} `json:"info"`
	Results []Episode `json:"results"`
}

func main() {
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/top-pairs", topPairsHandler)

	fmt.Println("Rick & Morty Backend running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Fatal: Could not start server: %s\n", err)
	}
}

// searchHandler hits three different API endpoints (chars, locations, episodes)
// at the same time
func searchHandler(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	limitStr := r.URL.Query().Get("limit")

	categories := []string{"character", "location", "episode"}
	var wg sync.WaitGroup
	resultsChan := make(chan []SearchResult, len(categories))

	for _, cat := range categories {
		wg.Add(1)
		go func(category string) {
			defer wg.Done()
			// If one category fails, we just get nil and keep going.
			res := fetchFromAPIWithURL("https://rickandmortyapi.com/api", category, term)
			resultsChan <- res
		}(cat)
	}

	// This goroutine ensures we dont block the main thread while waiting.
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var finalResults []SearchResult
	for res := range resultsChan {
		finalResults = append(finalResults, res...)
	}

	// Checking all pages.
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && len(finalResults) > limit {
			finalResults = finalResults[:limit]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(finalResults)
}

// Checks the pairs and counts them also if their occurences
// are swaped like Rick|Morty and Morty|Rick
func CountPairs(episodes []Episode) map[string]int {
	pairCounts := make(map[string]int)
	for _, ep := range episodes {
		chars := ep.Characters
		for i := 0; i < len(chars); i++ {
			for j := i + 1; j < len(chars); j++ {
				p1, p2 := chars[i], chars[j]
				if p1 > p2 {
					p1, p2 = p2, p1
				}
				pairCounts[p1+"|"+p2]++
			}
		}
	}
	return pairCounts
}

// topPairsHandler finds out which characters hang out together the most.
func topPairsHandler(w http.ResponseWriter, r *http.Request) {
	min, _ := strconv.Atoi(r.URL.Query().Get("min"))
	max, _ := strconv.Atoi(r.URL.Query().Get("max"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	// We need the NameMap because the Episode API only gives us URLs, not names.
	allEpisodes, nameMap := getAllData()

	pairCounts := CountPairs(allEpisodes)

	var pairs []PairResult
	for key, count := range pairCounts {
		// Filter based on user min and max frequency requirements.
		if (min > 0 && count < min) || (max > 0 && count > max) {
			continue
		}

		urls := strings.Split(key, "|")
		pairs = append(pairs, PairResult{
			Character1: CharacterInfo{Name: nameMap[urls[0]], URL: urls[0]},
			Character2: CharacterInfo{Name: nameMap[urls[1]], URL: urls[1]},
			Episodes:   count,
		})
	}

	// Show the most frequent pairs first.
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Episodes > pairs[j].Episodes
	})

	if len(pairs) > limit {
		pairs = pairs[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pairs)
}

// fetchFromAPI is a small helper for the search endpoint.
func fetchFromAPIWithURL(baseURL, category, term string) []SearchResult {
	url := fmt.Sprintf("%s/%s/?name=%s", baseURL, category, term)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var raw RawResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil
	}

	var formatted []SearchResult
	for _, item := range raw.Results {
		formatted = append(formatted, SearchResult{
			Name: item.Name,
			Type: category,
			URL:  item.URL,
		})
	}
	return formatted
}

// Using a bit of brute force we build a little local cache to store the data of episodes from API.
func getAllData() ([]Episode, map[string]string) {
	var all []Episode
	names := make(map[string]string)

	// Get every episode and assign to all slice
	nextURL := "https://rickandmortyapi.com/api/episode"
	for nextURL != "" {
		resp, err := http.Get(nextURL)
		if err != nil {
			break
		}

		var data EpisodeResponse
		json.NewDecoder(resp.Body).Decode(&data)
		all = append(all, data.Results...)
		nextURL = data.Info.Next
		resp.Body.Close()
	}

	// Get eery character name so we can connect URL to names.
	charURL := "https://rickandmortyapi.com/api/character"
	for charURL != "" {
		resp, err := http.Get(charURL)
		if err != nil {
			break
		}

		var data struct {
			Info    struct{ Next string } `json:"info"`
			Results []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"results"`
		}
		json.NewDecoder(resp.Body).Decode(&data)
		for _, c := range data.Results {
			names[c.URL] = c.Name
		}
		charURL = data.Info.Next
		resp.Body.Close()
	}

	return all, names
}
