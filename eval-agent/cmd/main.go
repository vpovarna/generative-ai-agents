package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/aggregator"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/config"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/prechecks"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/redis"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/stream"
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
		log.Warn().Msg("No .env file found")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize Bedrock client for embeddings
	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")
	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Bedrock client")
	}

	// Redis client
	streamCfg := stream.NewStreamConfig(
		os.Getenv("REDIS_ADDR"), // "localhost:6379"
		"eval-events",           // stream name
		"eval-group",            // consumer group
		os.Getenv("HOSTNAME"),   // unique consumer name
	)

	// Aggregator weights
	precheckWeight, err := strconv.ParseFloat(os.Getenv("PRECHECK_WEIGHT"), 64)
	if err != nil {
		precheckWeight = 0.3
	}
	llmJudgeWeight, err := strconv.ParseFloat(os.Getenv("LLM_JUDGE_WEIGHT"), 64)
	if err != nil {
		llmJudgeWeight = 0.7
	}

	redisClient, err := redis.ConnectRedis(ctx, os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"), 5)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	// Wire Components
	// Stage 1 — PreChecks
	stageRunner := prechecks.NewStageRunner([]prechecks.Checker{
		&prechecks.LengthChecker{},
		&prechecks.OverlapChecker{MinOverlapThreshold: 0.3},
		&prechecks.FormatChecker{},
	})

	// Stage 2 — LLM Judges (from YAML config)
	judgesConfig, err := config.LoadJudgesConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load judges config")
	}

	judgePool := judge.NewJudgePool(bedrockClient, &logger)
	judges, err := judgePool.BuildFromConfig(judgesConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build judges from config")
	}

	judgeRunner := judge.NewJudgeRunner(judges, &logger)

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

	consumer := stream.NewConsumer(redisClient, streamCfg.Stream, streamCfg.Group, streamCfg.ConsumerName, exec, &logger)

	// Setup consumer
	err = consumer.Setup(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup consumer")
	}

	// Start consumer
	go func() {
		if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error().Err(err).Msg("Consumer stopped with error")
		}
	}()

	// Wait for context to be done
	<-ctx.Done()
	logger.Info().Msg("Shutting down...")

	log.Info().Msg("Eval Agent stopped")
}
