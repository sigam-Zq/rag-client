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
