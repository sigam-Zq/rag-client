package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Request defines the structure for the Ollama embedding API request.
type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// Response defines the structure for the Ollama embedding API response.
type Response struct {
	Embedding []float64 `json:"embedding"`
}

// GetEmbedding sends text to the Ollama API and returns the embedding vector.
func GetEmbedding(text string) ([]float64, error) {
	reqBody := Request{
		Model:  "bge-m3", // Assuming bge-m3 model
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/embeddings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request to ollama failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("ollama error (status %d): %v", resp.StatusCode, errResp)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return result.Embedding, nil
}
