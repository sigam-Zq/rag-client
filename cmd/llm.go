/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"go-agents-cli/pkg/orchestrator"

	"github.com/spf13/cobra"
)

var (
	modelName string
	ragLimit  int
)

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm [question]",
	Short: "Ask a question to Ollama with RAG support",
	Long: `Retrieve relevant context from the vector database and use Ollama to generate a comprehensive answer.
Example: go-agents-cli llm "How to update CKA exam?" --model deepseek-r1:7b`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := strings.Join(args, " ")
		fmt.Printf("Querying with model: %s...\n", modelName)

		answer, err := orchestrator.Ask(question, modelName, ragLimit)
		if err != nil {
			return fmt.Errorf("failed to get answer: %w", err)
		}

		fmt.Printf("\n--- Answer ---\n%s\n", answer)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)

	// Add flags for model and limit
	llmCmd.Flags().StringVarP(&modelName, "model", "m", "llama3.1:8b", "Model name to use in Ollama (default: deepseek-r1:7b)")
	llmCmd.Flags().IntVarP(&ragLimit, "limit", "l", 5, "Number of context chunks to retrieve from Qdrant")
}
