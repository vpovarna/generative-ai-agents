package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/rewrite"
)

func main() {
	prompt := flag.String("prompt", "", "The prompt to send to Claude")
	stream := flag.Bool("stream", false, "Enable streaming response")
	maxTokens := flag.Int("max-tokens", 2000, "The maximum number of tokens to generate")
	stdin := flag.Bool("stdin", false, "Read prompt from stdin")

	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var finalPrompt string

	if *stdin {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal("Failed to read from stdin:", err)
		}
		finalPrompt = string(bytes)
	} else if *prompt != "" {
		finalPrompt = *prompt
	} else {
		log.Fatal("Please provide a prompt using -prompt or -stdin")
	}

	ctx := context.Background()

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)
	rewriter := rewrite.NewRewriter(bedrockClient)
	if err != nil {
		log.Fatal(err)
	}

	// Query rewrite
	rewrittenQuery, err := rewriter.RewriteQuery(ctx, finalPrompt)
	if err != nil {
		log.Println("Query rewrite failed")
		// Continue with original query
		rewrittenQuery = finalPrompt
	}

	req := bedrock.ClaudeRequest{
		Prompt:      rewrittenQuery,
		MaxTokens:   *maxTokens,
		Temperature: 0.0,
	}

	// Invoke client
	if *stream {
		fmt.Println("Streaming response:")
		response, err := bedrockClient.InvokeModelStream(ctx, req, func(chunk string) error {
			fmt.Print(chunk)
			return nil
		})
		if err != nil {
			log.Fatalf("Unable to invoke Claude model: %v", err)
		}
		fmt.Printf("\n\nStop reason: %s\n", response.StopReason)
	} else {
		response, err := bedrockClient.InvokeModel(ctx, req)
		if err != nil {
			log.Fatalf("Unable to invoke Claude model: %v", err)
		}
		fmt.Printf("Claude Response: \n%s\n", response.Content)
	}
}
