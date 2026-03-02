package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type Client struct {
	Client *bedrockruntime.Client
}

func NewClient(ctx context.Context, region string) (*Client, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))

	if err != nil {
		return nil, fmt.Errorf("Unable to load AWS config: %w", err)
	}

	// Create bedrockruntime client
	bedrockClient := bedrockruntime.NewFromConfig(cfg)

	// Return Bedrock Client
	return &Client{
		Client: bedrockClient,
	}, nil
}
