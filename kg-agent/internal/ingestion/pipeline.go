package ingestion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/embedding"
	"github.com/rs/zerolog/log"
)

type Pipeline struct {
	parser   *Parser
	chunker  *Chunker
	embedder *embedding.BedrockEmbedder
	pool     *pgxpool.Pool
}

func NewPipeline(
	parser *Parser,
	chunker *Chunker,
	embedder *embedding.BedrockEmbedder,
	pool *pgxpool.Pool,
) *Pipeline {
	return &Pipeline{
		parser:   parser,
		chunker:  chunker,
		embedder: embedder,
		pool:     pool,
	}
}

// IngestTextDocument processes a text file and stores it atomically
func (p *Pipeline) IngestTextDocument(ctx context.Context, filePath string) error {
	log.Info().Str("file", filePath).Msg("Starting ingestion")

	doc, err := p.parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("Failed to parse file. Error: %w", err)
	}
	log.Info().Str("doc_id", doc.ID).Str("title", doc.Title).Msg("Document parsed")

	chunks := p.chunker.TextChunker(doc.Content)
	log.Info().Int("chunk_count", len(chunks)).Msg("Document chunked successfully")

	var chunkContent []string
	for _, chunk := range chunks {
		chunkContent = append(chunkContent, chunk.Content)
	}

	embeddings, err := p.embedder.GenerateBatchEmbeddings(ctx, chunkContent)
	if err != nil {
		return fmt.Errorf("Failed to generate embeddings. Error: %w", err)
	}

	log.Info().Msg("Embeddings generated successfully")

	if err := p.storeDocumentWithChunks(ctx, doc, chunks, embeddings); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	log.Info().
		Str("doc_id", doc.ID).
		Int("chunks", len(chunks)).
		Msg("Ingestion complete")

	return nil
}

// InsertJsonDocument processes a json file and stores it atomically
func (p *Pipeline) IngestJsonDocument(ctx context.Context, filePath string) error {
	const BATCH_SIZE = 25

	log.Info().Str("file", filePath).Msg("Starting ingestion")

	doc, err := p.parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("Failed to parse file. Error: %w", err)
	}
	log.Info().Str("doc_id", doc.ID).Str("title", doc.Title).Msg("Document parsed")

	entries, err := p.chunker.JsonChunker(doc.Content)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := p.pool.Begin(ctx)

	if err := p.insertDocument(ctx, doc); err != nil {
		return err
	}

	for i := 0; i < len(entries); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(entries) {
			end = len(entries)
		}

		subset := entries[i:end]

		entriesContent := []string{}
		var chunks []Chunk
		for j, entry := range subset {
			entriesContent = append(entriesContent, entry.Content)
			entryMetadata := map[string]any{
				"question":    entry.Metadata.Question,
				"shortAnswer": entry.Metadata.ShortAnswer,
				"chunk_id":    entry.ChunkID,
			}
			chunks = append(chunks, Chunk{
				Index:    i + j,
				Start:    0,
				End:      len(entry.Content),
				Content:  entry.Content,
				Metadata: entryMetadata,
			})
		}
		batchEmbeddings, err := p.embedder.GenerateBatchEmbeddings(ctx, entriesContent)
		if err != nil {
			return err
		}

		if err := p.storeChunksOnly(ctx, doc.ID, chunks, batchEmbeddings); err != nil {
			return fmt.Errorf("failed to store chunks: %w", err)
		}

		log.Info().Int("batch", i/BATCH_SIZE+1).Int("chunks", len(chunks)).Msg("Batch complete")
	}

	// Commit EVERYTHING at once
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().Str("doc_id", doc.ID).Int("total_chunks", len(entries)).Msg("Ingestion complete")

	return nil
}

// storeDocumentWithChunks stores document and chunks in a single transaction
func (p *Pipeline) storeDocumentWithChunks(
	ctx context.Context,
	doc *Document,
	chunks []Chunk,
	embeddings [][]float32,
) error {
	// Begin transaction
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if we don't commit

	// Insert document
	docQuery := `
        INSERT INTO documents (id, title, content, metadata, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `
	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal document metadata: %w", err)
	}

	_, err = tx.Exec(ctx, docQuery, doc.ID, doc.Title, doc.Content, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	log.Info().Str("doc_id", doc.ID).Msg("Document inserted in transaction")

	// Insert all chunks
	chunkQuery := `
        INSERT INTO document_chunks (id, document_id, chunk_index, content, embedding, metadata, created_at)
        VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())
    `

	for i, chunk := range chunks {
		metadataJSON, _ := json.Marshal(chunk.Metadata)

		// Convert to pgvector.Vector
		vector := pgvector.NewVector(embeddings[i])

		_, err := tx.Exec(ctx, chunkQuery,
			doc.ID,        // document_id
			chunk.Index,   // chunk_index
			chunk.Content, // content
			vector,        // embedding
			metadataJSON,  // metadata
		)

		if err != nil {
			// Transaction will auto-rollback via defer
			return fmt.Errorf("failed to insert chunk %d: %w", i, err)
		}
	}
	log.Info().Int("chunks", len(chunks)).Msg("All chunks inserted in transaction")

	// Commit transaction - both document and chunks succeed together
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().Msg("Transaction committed successfully")
	return nil
}

// Add this new method to pipeline.go (before storeDocumentWithChunks)
func (p *Pipeline) insertDocument(ctx context.Context, doc *Document) error {
	docQuery := `
        INSERT INTO documents (id, title, content, metadata, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `
	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal document metadata: %w", err)
	}

	_, err = p.pool.Exec(ctx, docQuery, doc.ID, doc.Title, doc.Content, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	log.Info().Str("doc_id", doc.ID).Msg("Document inserted")
	return nil
}

// Add this new method to pipeline.go
func (p *Pipeline) storeChunksOnly(ctx context.Context, documentID string, chunks []Chunk, embeddings [][]float32) error {
	chunkQuery := `
        INSERT INTO document_chunks (id, document_id, chunk_index, content, embedding, metadata, created_at)
        VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())
    `

	for i, chunk := range chunks {
		metadataJSON, _ := json.Marshal(chunk.Metadata)
		vector := pgvector.NewVector(embeddings[i])

		_, err := p.pool.Exec(ctx, chunkQuery,
			documentID,    // document_id
			chunk.Index,   // chunk_index
			chunk.Content, // content
			vector,        // embedding
			metadataJSON,  // metadata
		)

		if err != nil {
			return fmt.Errorf("failed to insert chunk %d: %w", i, err)
		}
	}

	return nil
}
