package search

import (
	"context"
	"fmt"
	"sort"

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
				Metadata:   chunk.Metadata,
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
			Metadata:   chunk.Metadata,
			Rank:       i + 1,
		})
	}

	return searchResults, nil
}

type scoredResult struct {
	chunkID string
	score   float64
	result  SearchResult
}

func (s *Service) HybridSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Get result from semantic search and keyword search

	semanticResults, err := s.SemanticSearch(ctx, query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("Sematic search failed: %w", err)
	}

	keywordResults, err := s.KeywordSearch(ctx, query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("Keyword search failed: %w", err)
	}

	// Map will contain the rff scores. From chunk_id to total_rff_score
	rffScores := make(map[string]float64)
	// Map to store actual result object
	resultsMap := make(map[string]SearchResult)

	// RRF Formula: score = 1 / (k + rank); where k = 60 and rank is the position of the result
	k := 60.0

	// Process Semantic Search result
	for i, result := range semanticResults {
		rank := float64(i + 1) // is starting for 0
		rrfScore := 1.0 / (rank + k)

		// Add score to map (or create new entry)
		rffScores[result.ChunkID] += rrfScore

		// Store the result object
		resultsMap[result.ChunkID] = result
	}

	// Process keyword search results
	for i, result := range keywordResults {
		rank := float64(i + 1) // is starting for 0
		rrfScore := 1.0 / (rank + k)

		// Add to existing score (if chunk appeared in semantic too, the score will be higher)
		rffScores[result.ChunkID] += rrfScore
		// Only store if not already stored (semantic has priority)
		if _, exists := resultsMap[result.ChunkID]; !exists {
			resultsMap[result.ChunkID] = result
		}
	}

	//Key point: If a chunk appears in BOTH searches, it gets BOTH RRF scores added together!

	// Convert Map to Slice to sort based on rffScore
	var scored []scoredResult
	for chunkID, rffScore := range rffScores {
		scored = append(scored, scoredResult{
			chunkID: chunkID,
			score:   rffScore,
			result:  resultsMap[chunkID],
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score // Descendence
	})

	// Take top 'limit' results
	searchResults := []SearchResult{}
	for i := 0; i < limit && i < len(scored); i++ {
		searchResult := scored[i].result
		searchResult.Score = scored[i].score
		searchResult.Rank = i + 1
		searchResults = append(searchResults, searchResult)
	}

	return searchResults, nil

}

func (s *Service) DistanceToScore(distance float64) float64 {
	// Cosine distance range: 0 (identical) to 2 (opposite)
	// Convert to similarity score: 1 (best) to 0 (worst)
	score := 1.0 - distance

	// Clamp to [0, 1] range to avoid negative scores
	if score < 0.0 {
		return 0.0
	}
	if score > 1.0 {
		return 1.0
	}

	return score
}
