package ingestion

import (
	"encoding/json"
	"fmt"
)

type Chunker struct {
	ChunkSize    int
	ChunkOverlap int
}

type Chunk struct {
	Index    int
	Start    int
	End      int
	Content  string
	Metadata map[string]any
}

func NewChunker(chunkSize, overlap int) *Chunker {
	return &Chunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: overlap,
	}
}

func (c *Chunker) TextChunker(text string) []Chunk {
	// Validate chunk size and overlap
	if c.ChunkSize <= 0 || c.ChunkOverlap < 0 || c.ChunkOverlap >= c.ChunkSize {
		return []Chunk{}
	}

	results := []Chunk{}
	n := len(text)
	i := 0
	chunkIndex := 0

	for i < n {
		if i+c.ChunkSize <= n {
			acc := text[i : i+c.ChunkSize]

			chunk := Chunk{
				Index:    chunkIndex,
				Content:  acc,
				Start:    i,
				End:      i + c.ChunkSize,
				Metadata: map[string]any{},
			}
			results = append(results, chunk)
		} else {
			acc := text[i:]
			chunk := Chunk{
				Index:    chunkIndex,
				Content:  acc,
				Start:    i,
				End:      n,
				Metadata: map[string]any{},
			}

			results = append(results, chunk)
		}
		i = i + c.ChunkSize - c.ChunkOverlap
		chunkIndex++

	}

	return results
}

type Entry struct {
	ChunkID  string   `json:"chunk_id"`
	Content  string   `json:"content"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Question    string `json:"question"`
	ShortAnswer string `json:"short_answer"`
}

func (c *Chunker) JsonChunker(content string) ([]Entry, error) {
	var entries []Entry
	if err := json.Unmarshal([]byte(content), &entries); err != nil {
		return nil, fmt.Errorf("Unable to deserialize the input file: %w", err)
	}

	return entries, nil
}
