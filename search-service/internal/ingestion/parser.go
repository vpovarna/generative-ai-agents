package ingestion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Document struct {
	ID       string
	Title    string
	Content  string
	FilePath string
	Metadata map[string]string
}

type Parser struct {
	// Add parser specific configuration
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(path string) (*Document, error) {
	path = strings.TrimSpace(path)

	// Check file extension
	ext := filepath.Ext(path)

	// Read file content
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Check if file is empty
	if len(bytes) == 0 {
		return nil, fmt.Errorf("file %s is empty", path)
	}

	documentContent := string(bytes)

	filename := filepath.Base(path)
	title := strings.TrimSuffix(filename, ext)

	docID := uuid.New().String()

	return &Document{
		ID:       docID,
		Title:    title,
		Content:  documentContent,
		FilePath: path,
		Metadata: map[string]string{
			"filename":  filename,
			"extension": ext,
		},
	}, nil
}
