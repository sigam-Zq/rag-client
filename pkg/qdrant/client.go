package qdrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

// Point defines the structure for a single point to be inserted into Qdrant.
// Note: The 'ID' field can be a UUID string or an integer.
type Point struct {
	ID      interface{}            `json:"id"`
	Vector  []float64              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// InsertPoints inserts a batch of points into a specified Qdrant collection.
func InsertPoints(collectionName string, points []Point) error {
	if len(points) == 0 {
		return nil
	}

	if err := ensureCollectionExists(collectionName); err != nil {
		return fmt.Errorf("failed to ensure collection '%s' exists: %w", collectionName, err)
	}

	body := map[string]interface{}{
		"points": points,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal points: %w", err)
	}

	url := fmt.Sprintf("http://localhost:6333/collections/%s/points", collectionName)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request to qdrant failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("qdrant error (status %d): %v", resp.StatusCode, errResp)
	}

	return nil
}

// ensureCollectionExists checks if a collection exists and creates it if it doesn't.
func ensureCollectionExists(collectionName string) error {
	url := fmt.Sprintf("http://localhost:6333/collections/%s", collectionName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil // Collection already exists
	}

	// If not found (404), create it
	if resp.StatusCode == 404 {
		return createCollection(collectionName)
	}

	// Handle other unexpected status codes
	var errResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&errResp)
	return fmt.Errorf("unexpected error when checking collection (status %d): %v", resp.StatusCode, errResp)
}

// createCollection sends a request to create a new collection in Qdrant.
func createCollection(collectionName string) error {
	createBody := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     1024, // Assuming bge-m3 embedding size
			"distance": "Cosine",
		},
	}

	jsonData, err := json.Marshal(createBody)
	if err != nil {
		return fmt.Errorf("failed to marshal create collection request: %w", err)
	}

	url := fmt.Sprintf("http://localhost:6333/collections/%s", collectionName)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create new collection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("create collection request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("failed to create collection (status %d): %v", resp.StatusCode, errResp)
	}

	fmt.Printf("Collection '%s' created successfully.\n", collectionName)
	return nil
}

// SearchResult represents a single search result from Qdrant.
type SearchResult struct {
	Score          float64
	Payload        map[string]interface{}
	CollectionName string // To identify the source collection in global search
}

// Search performs a vector search in a specific Qdrant collection.
func Search(collectionName string, vec []float64, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	body := map[string]interface{}{
		"vector":       vec,
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	url := fmt.Sprintf("http://localhost:6333/collections/%s/points/search", collectionName)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("search request to qdrant failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("qdrant search error (status %d): %v", resp.StatusCode, errResp)
	}

	var result struct {
		Result []struct {
			Score   float64                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode qdrant search response: %w", err)
	}

	var results []SearchResult
	for _, r := range result.Result {
		results = append(results, SearchResult{
			Score:          r.Score,
			Payload:        r.Payload,
			CollectionName: collectionName,
		})
	}

	return results, nil
}

// ListCollections retrieves a list of all collection names from Qdrant.
func ListCollections() ([]string, error) {
	url := "http://localhost:6333/collections"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("qdrant list collections error (status %d): %v", resp.StatusCode, errResp)
	}

	var result struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode list collections response: %w", err)
	}

	var names []string
	for _, c := range result.Result.Collections {
		names = append(names, c.Name)
	}

	return names, nil
}

// SearchAllCollections performs a search across all collections and returns the top results.
func SearchAllCollections(vec []float64, limit int) ([]SearchResult, error) {
	collections, err := ListCollections()
	if err != nil {
		return nil, fmt.Errorf("could not get collection list for global search: %w", err)
	}

	var allResults []SearchResult
	// Create a channel to receive results from concurrent searches
	ch := make(chan []SearchResult)
	errCh := make(chan error)

	for _, name := range collections {
		go func(collectionName string) {
			results, err := Search(collectionName, vec, limit)
			if err != nil {
				errCh <- fmt.Errorf("search in collection '%s' failed: %w", collectionName, err)
				return
			}
			ch <- results
		}(name)
	}

	// Collect results
	for i := 0; i < len(collections); i++ {
		select {
		case results := <-ch:
			allResults = append(allResults, results...)
		case err := <-errCh:
			// For now, we just print the error and continue. You might want to handle this differently.
			fmt.Printf("Warning: %v\n", err)
		}
	}

	// Sort all results by score in descending order
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// Return only the top `limit` results
	if len(allResults) > limit {
		return allResults[:limit], nil
	}

	return allResults, nil
}
