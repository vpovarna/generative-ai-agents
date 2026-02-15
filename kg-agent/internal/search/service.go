package search

import (
	"context"
	"fmt"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/embedding"
)

type Service struct {
	db       *database.DB
	embedder *embedding.BedrockEmbedder
}

func NewService(db *database.DB, embedder *embedding.BedrockEmbedder) *Service {
	return &Service{
		db:       db,
		embedder: embedder,
	}
}

func (s *Service) SemanticSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	embeddings, err := s.embedder.GenerateEmbeddings(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate embeddings. Error: %w", err)
	}

	chunks, err := s.db.SemanticSearch(ctx, embeddings, limit)

	if err != nil {
		return nil, fmt.Errorf("Unable to run sematic search on the DB. Error: %w", err)
	}

	var searchResults []SearchResult

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

func (s *Service) KeywordSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	chunks, err := s.db.KeywordSearch(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("Unable to run keyword search on the DB. Error: %w", err)
	}

	var searchResults []SearchResult

	for i, chunk := range chunks {
		searchResults = append(searchResults, SearchResult{
			ChunkID:    chunk.Id,
			DocumentID: chunk.DocumentID,
			Content:    chunk.Content,
			Score:      chunk.Rank,
			Rank:       i + 1,
		})
	}

	return searchResults, nil
}

// func (s *SearchHandler) HybridSearch(ctx context.Context, query string, limit int) ([]SearchResult, error)

func (s *Service) DistanceToScore(distance float64) float64 {
	return 1.0 - distance
}
