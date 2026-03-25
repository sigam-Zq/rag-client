package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func GetEmbedding(text string) ([]float64, error) {
	reqBody := EmbeddingRequest{
		Model:  "bge-m3",
		Prompt: text,
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:11434/api/embeddings",
		"application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Embedding, nil
}

// 拆分文本
func SplitText(text string, chunkSize int) []string {
	var chunks []string
	runes := []rune(text)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

// 3️⃣ 写入 Qdrant
type QdrantPoint struct {
	ID      int                    `json:"id"`
	Vector  []float64              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

func InsertToQdrant(points []QdrantPoint) error {
	if len(points) == 0 {
		return nil
	}

	if err := ensureCollectionExists(); err != nil {
		return fmt.Errorf("failed to ensure collection exists: %w", err)
	}

	body := map[string]interface{}{
		"points": points,
	}

	jsonData, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT",
		"http://localhost:6333/collections/test_collection/points",
		bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("qdrant error %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

func ensureCollectionExists() error {
	resp, err := http.Get("http://localhost:6333/collections/test_collection")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	createBody := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     1024,
			"distance": "Cosine",
		},
	}

	jsonData, _ := json.Marshal(createBody)
	req, _ := http.NewRequest("PUT",
		"http://localhost:6333/collections/test_collection",
		bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("failed to create collection: %v", errResp)
	}

	return nil
}

// 4️⃣ 整体入库逻辑
func IndexDocument(doc string) error {
	chunks := SplitText(doc, 400)

	var points []QdrantPoint

	for i, chunk := range chunks {
		vec, err := GetEmbedding(chunk)
		if err != nil {
			return err
		}

		point := QdrantPoint{
			ID:     i,
			Vector: vec,
			Payload: map[string]interface{}{
				"text":   chunk,
				"doc_id": "doc1",
			},
		}

		points = append(points, point)
	}

	return InsertToQdrant(points)
}

type SearchResult struct {
	Score   float64
	Payload map[string]interface{}
}

func SearchQdrant(vec []float64, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	body := map[string]interface{}{
		"vector":       vec,
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
	}

	jsonData, _ := json.Marshal(body)

	resp, err := http.Post(
		"http://localhost:6333/collections/test_collection/points/search",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			Score   float64                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	var results []SearchResult
	for _, r := range result.Result {
		results = append(results, SearchResult{
			Score:   r.Score,
			Payload: r.Payload,
		})
	}

	return results, nil
}

// 3️⃣ 完整查询流程
func Query(query string) {
	vec, _ := GetEmbedding(query)
	results, _ := SearchQdrant(vec, 5)

	for _, r := range results {
		if text, ok := r.Payload["text"].(string); ok {
			fmt.Println(text)
		}
	}
}
