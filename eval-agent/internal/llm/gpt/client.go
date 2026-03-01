package gpt

import (
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Client struct {
	Client       openai.Client
	ModelID      string
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

func NewClient(apiKey string, model string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("OpenAI model ID is required")
	}

	openaiClient := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(3),
	)

	return &Client{
		Client:       openaiClient,
		ModelID:      model,
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     12 * time.Second,
	}, nil
}
