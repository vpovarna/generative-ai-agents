package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/agent"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/middleware"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/rewrite"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "KG Agent API",
			Description: "Knowledge Graph Agent with Claude",
			Version:     "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{
		{TagProps: spec.TagProps{Name: "health", Description: "Health checks"}},
		{TagProps: spec.TagProps{Name: "query", Description: "Query operations"}},
	}
}

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting KG Agent API Server")

	err := godotenv.Load()
	if err != nil {
		log.Error().Msg("No .env file found")
	}

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	ctx := context.Background()
	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to initialize Bedrock Client")
	}

	log.Info().
		Str("region", region).
		Str("model", modelID).
		Msg("Bedrock client initialized")

	rewriter := rewrite.NewRewriter(bedrockClient)
	handler := agent.NewHandler(bedrockClient, rewriter, modelID)

	container := restful.NewContainer()

	// Add filters
	container.Filter(middleware.Logger)
	container.Filter(middleware.RecoverPanic)

	// register API
	agent.RegisterRoutes(container, handler)

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		APIPath:                       "/api/v1/openapi.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}

	container.Add(restfulspec.NewOpenAPIService(config))

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	addr := fmt.Sprintf(":%s", port)
	log.Info().Str("address", addr).Msg("Starting server")

	server := http.Server{
		Addr:         addr,
		Handler:      corsHandler.Handler(container),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
