package orchestrator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"go-agents-cli/pkg/embedding"
	"go-agents-cli/pkg/qdrant"
	"go-agents-cli/pkg/util"

	"github.com/google/uuid"
)

// IndexDocument splits a document, generates embeddings, and indexes it into a dynamically named collection.
func IndexDocument(filePath, docContent string) error {
	// Sanitize file path to create a valid collection name
	collectionName := SanitizeCollectionName(filePath)
	docID := uuid.New().String()

	chunks := util.SplitText(docContent, 400)

	var points []qdrant.Point

	for i, chunk := range chunks {
		vec, err := embedding.GetEmbedding(chunk)
		if err != nil {
			return fmt.Errorf("error getting embedding for chunk %d: %w", i, err)
		}

		point := qdrant.Point{
			ID:     uuid.New().String(), // Each point gets a unique ID
			Vector: vec,
			Payload: map[string]interface{}{
				"text":   chunk,
				"doc_id": docID, // Associate all chunks with the same document
				"source": filepath.Base(filePath),
			},
		}

		points = append(points, point)
	}

	return qdrant.InsertPoints(collectionName, points)
}

// SanitizeCollectionName creates a valid Qdrant collection name from a file path.
// It removes the extension, converts to lowercase, and replaces invalid characters with underscores.
func SanitizeCollectionName(filePath string) string {
	// Get the base name without extension
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// Convert to lowercase
	name := strings.ToLower(baseName)

	// Qdrant collection names must match ^[a-zA-Z0-9_-]+$
	// Replace any invalid characters with an underscore
	reg := regexp.MustCompile(`[^a-z0-9_-]+`)
	name = reg.ReplaceAllString(name, "_")

	// Ensure it doesn't start or end with an underscore
	name = strings.Trim(name, "_")

	if name == "" {
		return "default_collection"
	}

	return name
}

// Ask takes a question, searches for relevant context in the vector database, and uses an LLM to generate an answer.
func Ask(question, model string, limit int) (string, error) {
	// 1. Get embedding for the question
	vec, err := embedding.GetEmbedding(question)
	if err != nil {
		return "", fmt.Errorf("failed to get embedding for question: %w", err)
	}

	// 2. Search across all collections in Qdrant
	results, err := qdrant.SearchAllCollections(vec, limit)
	if err != nil {
		return "", fmt.Errorf("failed to search collections: %w", err)
	}

	if len(results) == 0 {
		// If no context found, still ask the LLM but inform it there's no specific context.
		messages := []embedding.ChatMessage{
			{
				Role:    "user",
				Content: question,
			},
		}
		return embedding.Chat(model, messages)
	}

	// 3. Format the retrieved context
	var contextBuilder strings.Builder
	for i, r := range results {
		if text, ok := r.Payload["text"].(string); ok {
			contextBuilder.WriteString(fmt.Sprintf("[Context %d (from %s)]: %s\n", i+1, r.CollectionName, text))
		}
	}

	// 4. Construct a prompt with context and question
	prompt := fmt.Sprintf(`Use the following retrieved context to answer the user's question. If the context doesn't contain enough information, use your general knowledge but mention that the context was insufficient.

Context:
%s

Question: %s`, contextBuilder.String(), question)

	messages := []embedding.ChatMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 5. Call LLM
	return embedding.Chat(model, messages)
}
