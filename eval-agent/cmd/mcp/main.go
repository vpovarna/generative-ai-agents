package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/aggregator"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/mcpadapter"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/prechecks"
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
	exec := executor.NewExecutor(stageRunner, judgeRunner, agg, earlyExit, &logger)

	// Create MCP Server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "eval-agent",
			Version: "1.0.0",
		}, nil,
	)

	// Add Tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "evaluate_response",
		Description: "Evaluate an AI agent response for relevance, faithfulness, coherence, completeness, and instruction-following",
	}, mcpadapter.NewEvaluateHandler(exec))

	// Run over stdio
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		logger.Error().Err(err).Msg("Failed to run mcp server")
		os.Exit(1)
	}
}
