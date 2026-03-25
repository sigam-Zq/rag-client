/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"go-agents-cli/pkg/embedding"
	"go-agents-cli/pkg/qdrant"

	"github.com/spf13/cobra"
)

var (
	query          string
	limit          int
	showScore      bool
	collectionName string
)

func init() {
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "search query")
	searchCmd.Flags().IntVarP(&limit, "limit", "n", 5, "number of results to return")
	searchCmd.Flags().BoolVar(&showScore, "score", false, "show similarity score")
	searchCmd.Flags().StringVarP(&collectionName, "collection", "c", "", "specify a collection to search in (default: search all)")
	searchCmd.MarkFlagRequired("query")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search the vector database",
	Long:  `Query the vector database for matching documents. By default, it searches all collections.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vec, err := embedding.GetEmbedding(query)
		if err != nil {
			return fmt.Errorf("failed to get embedding: %w", err)
		}

		var results []qdrant.SearchResult
		if collectionName != "" {
			// Search in a specific collection
			fmt.Printf("Searching in collection: %s...\n", collectionName)
			results, err = qdrant.Search(collectionName, vec, limit)
		} else {
			// Search in all collections
			fmt.Println("Searching in all collections...")
			results, err = qdrant.SearchAllCollections(vec, limit)
		}

		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("\nFound %d results:\n", len(results))
		for i, r := range results {
			fmt.Printf("--- Result %d ---\n", i+1)
			if showScore {
				fmt.Printf("Score: %.4f\n", r.Score)
			}
			if r.CollectionName != "" {
				fmt.Printf("Collection: %s\n", r.CollectionName)
			}
			if text, ok := r.Payload["text"].(string); ok {
				fmt.Printf("Text: %s\n", strings.TrimSpace(text))
			}
			if source, ok := r.Payload["source"].(string); ok {
				fmt.Printf("Source: %s\n", source)
			}
			if docID, ok := r.Payload["doc_id"].(string); ok {
				fmt.Printf("Doc ID: %s\n", docID)
			}
		}
		return nil
	},
}
