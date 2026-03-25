package util

import (
	"fmt"
	"os"
	"strings"

	"baliance.com/gooxml/document"
	pdf "github.com/ledongthuc/pdf"
)

// ReadTextFile reads a plain text file (e.g., .md, .txt)
func ReadTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

// ReadDocxFile extracts text from a .docx file using a robust library.
func ReadDocxFile(path string) (string, error) {
	doc, err := document.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening docx file: %w", err)
	}

	var builder strings.Builder
	for _, p := range doc.Paragraphs() {
		for _, r := range p.Runs() {
			builder.WriteString(r.Text())
		}
		builder.WriteString("\n") // Add a newline after each paragraph
	}
	return builder.String(), nil
}

// ReadPdfFile extracts text from a .pdf file
func ReadPdfFile(path string) (string, error) {
	return ReadPdfLib(path)
}

func ReadPdfLib(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf strings.Builder
	totalPages := r.NumPage()

	for i := 1; i <= totalPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
