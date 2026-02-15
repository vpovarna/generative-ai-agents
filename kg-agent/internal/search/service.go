package search

import (
	"context"
	"fmt"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/embedding"
)

type SearchHandler struct {
	db       *database.DB
	embedder *embedding.BedrockEmbedder
}

func NewHandler(db *database.DB, embedder *embedding.BedrockEmbedder) *SearchHandler {
	return &SearchHandler{
		db:       db,
		embedder: embedder,
	}
}

func (s *SearchHandler) SematicSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	embeddings, err := s.embedder.GenerateEmbeddings(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate embeddings. Error: %w", err)
	}

	chunks, err := s.db.SemanticSearch(ctx, embeddings, limit)

	if err != nil {
		return nil, fmt.Errorf("Unable to run sematic search on the DB. Error: %w", err)
	}

	searchResults := []SearchResult{}

	for i, chunk := range chunks {
		searchResults = append(searchResults,
			SearchResult{
				ChunkID:    chunk.Id,
				DocumentID: chunk.DocumentID,
				Content:    chunk.Content,
				Score:      s.DistanceToScore(chunk.Distance),
				Rank:       i + 1, // Position of the chunk in the result
			})
	}

	return searchResults, nil
}

func (s *SearchHandler) KeywordSearch(ctx context.Context, query string, limit int) ([]SearchResult, error)

func (s *SearchHandler) HybridSearch(ctx context.Context, query string, limit int) ([]SearchResult, error)

func (s *SearchHandler) DistanceToScore(distance float64) float64 {
	return 1.0 - distance
}
