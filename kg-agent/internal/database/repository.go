package database

import (
	"context"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
)

func (db *DB) DeleteDocument(ctx context.Context, docId string) error {

	query := `DELETE FROM documents WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, docId)
	if err != nil {
		return fmt.Errorf("Failed to delete document id: %s, error: %w", docId, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Warn().Str("doc_id", docId).Msg("Document not found")
	} else {
		log.Info().Str("doc_id", docId).Msg("Document deleted")
	}

	return nil
}

// TODO: Add pagination
func (db *DB) GetAllDocs(ctx context.Context) ([]Document, error) {
	query := `SELECT id, title from documents`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch document ids from DB")
	}

	defer rows.Close()

	var documentsResponse []Document

	for rows.Next() {
		var document Document

		if err := rows.Scan(&document.Id, &document.Title); err != nil {
			return nil, fmt.Errorf("Failed to scan id: %w", err)
		}

		documentsResponse = append(documentsResponse, document)
	}

	return documentsResponse, nil
}

// TODO: Add support for cosine and euclidean distance configuration
func (db *DB) SemanticSearch(ctx context.Context, queryEmbeddings []float32, limit int) ([]Chunk, error) {
	// Convert embeddings to pgvector embeddings
	pgvectorEmbeddings := pgvector.NewVector(queryEmbeddings)

	query := `
	SELECT 
	  id, 
	  document_id,
	  content, 
	  metadata,
	  embedding <=> $1 AS distance 
	FROM document_chunks 
	ORDER BY distance ASC 
	LIMIT $2`

	rows, err := db.Pool.Query(ctx, query, pgvectorEmbeddings, limit)

	if err != nil {
		return nil, fmt.Errorf("Unable to query the database: %w", err)
	}

	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var chunk Chunk

		if err := rows.Scan(&chunk.Id, &chunk.DocumentID, &chunk.Content, &chunk.Metadata, &chunk.Distance); err != nil {
			return nil, fmt.Errorf("Failed to scan id: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	// Rows errors catch
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return chunks, nil
}

func (db *DB) KeywordSearch(ctx context.Context, userQuery string, limit int) ([]Chunk, error) {
	query := `
		SELECT 
			id,
			document_id,
			content,
			metadata,
			ts_rank(content_tsvector, plainto_tsquery('english', $1)) AS rank
		FROM document_chunks
		WHERE content_tsvector @@ plainto_tsquery('english', $1)
		ORDER BY rank DESC
		LIMIT $2`

	rows, err := db.Pool.Query(ctx, query, userQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed. Error: %w", err)
	}

	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var chunk Chunk

		err := rows.Scan(&chunk.Id, &chunk.DocumentID, &chunk.Content, &chunk.Metadata, &chunk.Rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	// Rows errors catch
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return chunks, nil
}
