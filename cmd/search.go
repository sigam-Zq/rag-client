/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"go-agents-cli/pkg"

	"github.com/spf13/cobra"
)

var (
	query     string
	limit     int
	showScore bool
)

func init() {
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "search query")
	searchCmd.Flags().IntVarP(&limit, "limit", "n", 5, "number of results to return")
	searchCmd.Flags().BoolVar(&showScore, "score", false, "show similarity score")
	searchCmd.MarkFlagRequired("query")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search the vector database",
	Long:  `Query the vector database for matching documents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vec, err := pkg.GetEmbedding(query)
		if err != nil {
			return fmt.Errorf("failed to get embedding: %w", err)
		}

		results, err := pkg.SearchQdrant(vec, limit)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		for i, r := range results {
			fmt.Printf("--- Result %d ---\n", i+1)
			if showScore {
				fmt.Printf("Score: %.4f\n", r.Score)
			}
			if text, ok := r.Payload["text"].(string); ok {
				fmt.Printf("%s\n", strings.TrimSpace(text))
			}
			if docID, ok := r.Payload["doc_id"].(string); ok {
				fmt.Printf("Doc ID: %s\n", docID)
			}
		}
		return nil
	},
}
