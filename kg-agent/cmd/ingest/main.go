package main

import (
	"context"
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/embedding"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/ingestion"
	"github.com/rs/zerolog/log"
)

func main() {
	filePath := flag.String("filePath", "resources/test-input.txt", "Relative path to the document")
	chunkSize := flag.Int("chunkSize", 500, "Chunk size")
	chunkOverlap := flag.Int("chunkOverlap", 100, "Chunk overlap")

	flag.Parse()

	err := godotenv.Load()

	if err != nil {
		log.Warn().Msg("Unable to load env variables")
	}

	ctx := context.Background()

	config := database.Config{
		Host:     os.Getenv("KG_AGENT_VECTOR_DB_HOST"),
		Port:     os.Getenv("KG_AGENT_VECTOR_DB_PORT"),
		User:     os.Getenv("KG_AGENT_VECTOR_DB_USER"),
		Password: os.Getenv("KG_AGENT_VECTOR_DB_PASSWORD"),
		Database: os.Getenv("KG_AGENT_VECTOR_DB_DATABASE"),
		SSLMode:  os.Getenv("KG_AGENT_VECTOR_DB_SSLMode"),
	}

	db, err := database.NewWithBackoff(ctx, config, 3)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return
	}

	defer db.Close()

	log.Info().Msg("Database connected")

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)

	if err != nil {
		log.Error().Err(err).Msg("Unable to create bedrock client")
		return
	}

	parser := ingestion.NewParser()
	chunker := ingestion.NewChunker(*chunkSize, *chunkOverlap)
	embedder := embedding.NewBedrockEmbedder(bedrockClient.Client)

	// Create pipeline
	pipeline := ingestion.NewPipeline(parser, chunker, embedder, db.Pool)

	// Ingest document (atomic operation)
	if err := pipeline.IngestDocument(ctx, *filePath); err != nil {
		log.Fatal().Err(err).Msg("Ingestion failed")
	}

	log.Info().Msg("Ingestion successful!")
}
