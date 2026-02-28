package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/agent"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/cache"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/conversation"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/guardrails"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/middleware"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/redis"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/rewrite"
	"github.com/povarna/generative-ai-agents/kg-agent/internal/strategy"
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
	miniModelID := os.Getenv("CLAUDE_MINI_MODEL_ID")
	port := os.Getenv("AGENT_API_PORT")
	if port == "" {
		port = "8081"
	}

	searchApiBaseUrl := os.Getenv("SEARCH_API_URL")

	if searchApiBaseUrl == "" {
		searchApiBaseUrl = "http://localhost:8082"
	}
	searchApiTimeout := os.Getenv("SEARCH_API_TIMEOUT")
	timeout, err := strconv.Atoi(searchApiTimeout)
	if err != nil || searchApiTimeout == "" {
		timeout = 10
	}

	searchConfig := agent.SearchClientConfig{
		BaseURL:             searchApiBaseUrl,
		Timeout:             time.Duration(timeout) * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}

	ctx := context.Background()
	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to initialize Bedrock Client")
	}
	miniClient, err := bedrock.NewClient(ctx, region, miniModelID)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to initialize Mini Bedrock Client")
	}

	log.Info().
		Str("region", region).
		Str("model", modelID).
		Msg("Bedrock client initialized")

	// Connect to Redis with retries
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := redis.ConnectRedis(
		ctx,
		redisAddr,
		os.Getenv("REDIS_PASSWORD"),
		5, // max retries
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	configuredRedisTTL := os.Getenv("REDIS_TTL")
	var redisTTL time.Duration
	if configuredRedisTTL == "" {
		redisTTL = 30 * time.Minute
	} else {
		ttl, err := time.ParseDuration(configuredRedisTTL)
		if err != nil {
			redisTTL = 30 * time.Minute
		} else {
			redisTTL = ttl
		}
	}

	guardrailsValidator := guardrails.NewGuardrails(miniClient)
	rewriter := rewrite.NewRewriter(miniClient)
	searchClient := agent.NewSearchClient(searchConfig)
	retrievalStrategy := strategy.NewRetrievalStrategy(miniClient)
	conversationStore := conversation.NewRedisConversationStore(redisClient, redisTTL)
	searchCache := cache.NewRedisSearchCache(redisClient, "search_cache:")
	service := agent.NewService(
		bedrockClient,
		miniClient,
		modelID,
		rewriter,
		searchClient,
		conversationStore,
		retrievalStrategy,
		searchCache,
	)
	handler := agent.NewHandler(service, guardrailsValidator)

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
