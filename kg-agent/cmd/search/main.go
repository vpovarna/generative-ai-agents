package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/database"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/embedding"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/search"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("No .env file found")
	}

	ctx := context.Background()

	// Connect to database
	config := database.Config{
		Host:     os.Getenv("KG_AGENT_VECTOR_DB_HOST"),
		Port:     os.Getenv("KG_AGENT_VECTOR_DB_PORT"),
		User:     os.Getenv("KG_AGENT_VECTOR_DB_USER"),
		Password: os.Getenv("KG_AGENT_VECTOR_DB_PASSWORD"),
		Database: os.Getenv("KG_AGENT_VECTOR_DB_DATABASE"),
		SSLMode:  os.Getenv("KG_AGENT_VECTOR_DB_SSLMode"),
	}

	db, err := database.NewWithBackoff(ctx, config, 5)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Database connected")

	// Initialize Bedrock client for embeddings
	region := os.Getenv("AWS_REGION")
	bedrockClient, err := bedrock.NewClient(ctx, region, "")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Bedrock client")
	}

	// Wire Components
	embedder := embedding.NewBedrockEmbedder(bedrockClient.Client)
	searchService := search.NewService(db, embedder)
	handler := search.NewSearchHandler(searchService)

	// Setup routes
	container := restful.NewContainer()
	search.RegisterRoutes(container, handler)

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Start server
	port := os.Getenv("SEARCH_API_PORT")
	if port == "" {
		port = "8082"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("address", addr).Msg("Starting Search API")

	server := http.Server{
		Addr:    addr,
		Handler: corsHandler.Handler(container),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
