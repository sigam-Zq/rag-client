/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pdf "github.com/ledongthuc/pdf"

	"go-agents-cli/pkg"

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
		case ".md":
			content, err = readTextFile(filePath)
		case ".docx":
			content, err = readDocxFile(filePath)
		case ".pdf":
			content, err = readPdfFile(filePath)
		default:
			return fmt.Errorf("unsupported file format: %s (supported: .md, .docx, .pdf)", ext)
		}

		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if content == "" {
			return fmt.Errorf("file is empty or could not extract content")
		}

		baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		tmpFile := filepath.Join(".", baseName+"_extracted.txt")
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write tmp file: %v\n", err)
		}

		return pkg.IndexDocument(content)
	},
}

func readTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

func readDocxFile(path string) (string, error) {
	reader, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	doc, err := readDocx(reader)
	if err != nil {
		return "", err
	}
	return doc, nil
}

func readPdfFile(path string) (string, error) {
	return readPdfLib(path)
}

type docxReader struct{}

func readDocx(r io.ReaderAt) (string, error) {
	doc := NewDocxReader(r)
	paragraphs, err := doc.Read()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for _, p := range paragraphs {
		builder.WriteString(p)
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func readPdf(r io.ReaderAt) (string, error) {
	pdfReader, err := NewPdfReader(r)
	if err != nil {
		return "", err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			continue
		}
		content, err := page.GetContent()
		if err != nil {
			continue
		}
		builder.WriteString(content)
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

type DocxReader struct {
	r io.ReaderAt
}

func NewDocxReader(r io.ReaderAt) *DocxReader {
	return &DocxReader{r: r}
}

func (d *DocxReader) Read() ([]string, error) {
	return extractDocxText(d.r)
}

func extractDocxText(r io.ReaderAt) ([]string, error) {
	zipReader, err := NewZipReader(r)
	if err != nil {
		return nil, err
	}

	for _, file := range zipReader.ListFiles() {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			return parseDocxXml(string(data)), nil
		}
	}
	return nil, fmt.Errorf("document.xml not found in docx")
}

func parseDocxXml(xml string) []string {
	var paragraphs []string
	lines := strings.Split(xml, "<w:p ")
	for _, line := range lines {
		start := strings.Index(line, ">")
		if start == -1 {
			continue
		}
		line = line[start+1:]

		end := strings.Index(line, "</w:p>")
		if end == -1 {
			continue
		}
		line = line[:end]

		text := extractDocxTextFromElement(line)
		if text != "" {
			paragraphs = append(paragraphs, text)
		}
	}
	return paragraphs
}

func extractDocxTextFromElement(xml string) string {
	var result strings.Builder
	remaining := xml
	for {
		start := strings.Index(remaining, "<w:t")
		if start == -1 {
			break
		}
		remaining = remaining[start:]
		endTag := strings.Index(remaining, ">")
		if endTag == -1 {
			break
		}
		remaining = remaining[endTag+1:]
		end := strings.Index(remaining, "</w:t>")
		if end == -1 {
			break
		}
		result.WriteString(strings.TrimSpace(remaining[:end]))
		remaining = remaining[end+6:]
	}
	return strings.TrimSpace(result.String())
}

type ZipReader struct {
	r    io.ReaderAt
	size int64
}

func NewZipReader(r io.ReaderAt) (*ZipReader, error) {
	return &ZipReader{r: r}, nil
}

func (z *ZipReader) ListFiles() []ZipFile {
	var files []ZipFile
	header := make([]byte, 4)
	z.r.ReadAt(header, 0)
	if string(header) != "PK\x03\x04" {
		return files
	}

	var offset int64 = 0
	for {
		header := make([]byte, 30)
		n, err := z.r.ReadAt(header, offset)
		if n == 0 || err != nil {
			break
		}

		if string(header[0:4]) != "PK\x03\x04" {
			break
		}

		compSize := int64(header[18]) | int64(header[19])<<8 | int64(header[20])<<16 | int64(header[21])<<24
		compMethod := int64(header[8]) | int64(header[9])<<8
		nameLen := int64(header[26]) | int64(header[27])<<8
		extraLen := int64(header[28]) | int64(header[29])<<8

		nameBytes := make([]byte, nameLen)
		z.r.ReadAt(nameBytes, offset+30)

		if string(header[0:4]) == "PK\x01\x02" {
			break
		}

		files = append(files, ZipFile{
			Name:       string(nameBytes),
			Offset:     offset + 30 + nameLen + extraLen,
			CompSize:   compSize,
			CompMethod: int(compMethod),
			R:          z.r,
		})

		offset += 30 + nameLen + extraLen + compSize
	}
	return files
}

type ZipFile struct {
	Name       string
	Offset     int64
	CompSize   int64
	CompMethod int
	R          io.ReaderAt
}

func (z *ZipFile) Open() (io.ReadCloser, error) {
	data := make([]byte, z.CompSize)
	z.R.ReadAt(data, z.Offset)
	return io.NopCloser(strings.NewReader(string(data))), nil
}

type PdfReader struct {
	r    io.ReaderAt
	data []byte
}

func NewPdfReader(r io.ReaderAt) (*PdfReader, error) {
	size := int64(0)
	if f, ok := r.(interface{ Stat() (os.FileInfo, error) }); ok {
		if info, err := f.Stat(); err == nil {
			size = info.Size()
		}
	}

	if size == 0 {
		size = 1024 * 1024
	}

	data := make([]byte, size)
	n, _ := r.ReadAt(data, 0)
	if n < len(data) {
		data = data[:n]
	}
	return &PdfReader{r: r, data: data}, nil
}

func (p *PdfReader) GetNumPages() (int, error) {
	content := string(p.data)
	count := strings.Count(content, "/Type /Page")
	if count == 0 {
		count = strings.Count(content, "endobj")
	}
	return count, nil
}

func (p *PdfReader) GetPage(num int) (*PdfPage, error) {
	return &PdfPage{data: string(p.data), pageNum: num}, nil
}

type PdfPage struct {
	data    string
	pageNum int
}

func (p *PdfPage) GetContent() (string, error) {
	return extractTextFromPdfPage(p.data), nil
}

func extractTextFromPdfPage(data string) string {
	var result strings.Builder
	lines := strings.Split(data, "BT")
	for i := 1; i < len(lines); i++ {
		texts := strings.Split(lines[i], "ET")
		if len(texts) > 0 {
			content := texts[0]
			start := 0
			for {
				tj := strings.Index(content[start:], "Tj")
				td := strings.Index(content[start:], "TD")
				if tj == -1 && td == -1 {
					break
				}
				if tj != -1 && (td == -1 || tj < td) {
					start += tj
					end := strings.Index(content[start:], ")")
					if end != -1 && end < 20 {
						text := content[start+1 : start+end]
						text = strings.Trim(text, " ()[]<>/")
						if text != "" {
							result.WriteString(text)
						}
					}
					start += 2
				} else {
					if td != -1 {
						start += td + 2
					} else {
						break
					}
				}
			}
			result.WriteString("\n")
		}
	}
	return result.String()
}

func readPdfLib(filePath string) (string, error) {
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
