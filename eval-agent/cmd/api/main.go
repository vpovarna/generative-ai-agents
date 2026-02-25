package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/aggregator"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/api"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/api/middleware"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/prechecks"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	logger := log.Logger

	// Load env
	if err := godotenv.Load(); err != nil {
		logger.Warn().Msg("No .env file found")
	}

	ctx := context.Background()

	// Initialize Bedrock client for embeddings
	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	// Aggregator weights
	precheckWeight, err := strconv.ParseFloat(os.Getenv("PRECHECK_WEIGHT"), 64)
	if err != nil {
		precheckWeight = 0.3
	}
	llmJudgeWeight, err := strconv.ParseFloat(os.Getenv("LLM_JUDGE_WEIGHT"), 64)
	if err != nil {
		llmJudgeWeight = 0.7
	}

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Bedrock client")
	}

	// Wire Components
	// Stage 1 — PreChecks
	stageRunner := prechecks.NewStageRunner([]prechecks.Checker{
		&prechecks.LengthChecker{},
		&prechecks.OverlapChecker{MinOverlapThreshold: 0.3},
		&prechecks.FormatChecker{},
	})
	// Stage 2 — LLM Judges
	judgeRunner := judge.NewJudgeRunner([]judge.Judge{
		judge.NewRelevanceJudge(bedrockClient, &logger),
		judge.NewCoherenceJudge(bedrockClient, &logger),
		judge.NewFaithfulnessJudge(bedrockClient, &logger),
		judge.NewCompletenessJudge(bedrockClient, &logger),
		judge.NewInstructionJudge(bedrockClient, &logger),
	}, &logger)

	judges := judge.NewJudgeFactory(bedrockClient, &logger)

	// Aggregator
	agg := aggregator.NewAggregator(aggregator.Weights{
		PreChecks: precheckWeight,
		LLMJudge:  llmJudgeWeight,
	}, &logger)

	// Executor
	earlyExit, _ := strconv.ParseFloat(os.Getenv("EARLY_EXIT_THRESHOLD"), 64)
	if earlyExit == 0 {
		earlyExit = 0.2
	}
	agentExecutor := executor.NewExecutor(stageRunner, judgeRunner, agg, earlyExit, &logger)
	judgeExecutor := executor.NewJudgeExecutor(judges, &logger)

	// API
	handler := api.NewHandler(agentExecutor, judgeExecutor, &logger)
	container := restful.NewContainer()
	container.Filter(middleware.Logger)
	container.Filter(middleware.RecoverPanic)
	api.RegisterRoutes(container, handler)

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Server
	port := os.Getenv("EVAL_AGENT_API_PORT")
	if port == "" {
		port = "18081"
	}

	addr := fmt.Sprintf(":%s", port)
	logger.Info().Str("address", addr).Msg("Starting Eval Agent API")

	server := http.Server{
		Addr:    addr,
		Handler: corsHandler.Handler(container),
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal().Err(err).Msg("Server failed")
	}
}
