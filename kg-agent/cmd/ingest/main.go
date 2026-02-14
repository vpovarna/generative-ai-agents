package main

import (
	"context"
	"flag"
	"fmt"
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

	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Database ping failed")
		return
	}

	log.Info().Msg("Connected successfully")

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)

	if err != nil {
		log.Error().Err(err).Msg("Unable to create bedrock client")
		return
	}

	embedder := embedding.NewBedrockEmbedder(bedrockClient.Client)

	embedding, err := embedder.GenerateEmbeddings(ctx, "Hello world")
	if err != nil {
		log.Error().Err(err).Msg("Unable to generate embeddings")
		return
	}

	log.Info().
		Int("dimensions", len(embedding)).
		Float32("embeddings", embedding[0]).
		Msg("Embedding generated")

	// Test batch
	texts := []string{"Hello", "World", "Go programming"}
	embeddings, err := embedder.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		log.Error().Err(err).Msg("Unable to generate embeddings")
		return
	}

	log.Info().
		Int("count", len(embeddings)).
		Msg("Batch embeddings generated")

	parser := ingestion.NewParser()
	doc, err := parser.ParseFile(*filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse input file.")
	}

	chunker := ingestion.NewChunker(50, 10)
	chunks := chunker.ChunkText(doc.Content)

	for _, chunk := range chunks {
		fmt.Println(chunk)
	}

}
