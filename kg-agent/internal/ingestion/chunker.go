package ingestion

type Chunker struct {
	ChunkSize    int
	ChunkOverlap int
}

type Chunk struct {
	Index   int
	Start   int
	End     int
	Content string
}

func NewChunker(chunkSize, overlap int) *Chunker {
	return &Chunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: overlap,
	}
}

func (c *Chunker) ChunkText(text string) []Chunk {
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
				Index:   chunkIndex,
				Content: acc,
				Start:   i,
				End:     i + c.ChunkSize,
			}
			results = append(results, chunk)
		} else {
			acc := text[i:]
			chunk := Chunk{
				Index:   chunkIndex,
				Content: acc,
				Start:   i,
				End:     n,
			}

			results = append(results, chunk)
		}
		i = i + c.ChunkSize - c.ChunkOverlap
		chunkIndex++

	}

	return results
}
