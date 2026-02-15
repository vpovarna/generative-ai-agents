package search

type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty" description:"Max results(default: 10)"`
}

type SearchResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
}

type SearchResponse struct {
	Query  string         `json:"query"`
	Result []SearchResult `json:"result"`
	Count  int            `json:"count"`
	Method string         `json:"method"` // "semantic", "keyword", "hybrid"
}
