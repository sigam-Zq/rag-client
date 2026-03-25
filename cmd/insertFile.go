/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"go-agents-cli/pkg/orchestrator"
	"go-agents-cli/pkg/util"

	"github.com/spf13/cobra"
)

var (
	filePath string
)

func init() {
	insertFile.Flags().StringVarP(&filePath, "file", "f", "", "Path to the document file (docx, md, pdf)")
	insertFile.MarkFlagRequired("file")
	rootCmd.AddCommand(insertFile)
}

var insertFile = &cobra.Command{
	Use:   "InsertFile",
	Short: "Insert a document into vector database",
	Long:  `Read and index a document (docx, md, pdf) into the vector database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ext := strings.ToLower(filepath.Ext(filePath))
		var content string
		var err error

		switch ext {
		case ".md", ".txt":
			content, err = util.ReadTextFile(filePath)
		case ".docx":
			content, err = util.ReadDocxFile(filePath)
		case ".pdf":
			content, err = util.ReadPdfFile(filePath)
		default:
			return fmt.Errorf("unsupported file format: %s (supported: .md, .txt, .docx, .pdf)", ext)
		}

		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if content == "" {
			return fmt.Errorf("file is empty or could not extract content")
		}

		return orchestrator.IndexDocument(filePath, content)
	},
}
