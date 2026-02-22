package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type Client struct {
	Client  *bedrockruntime.Client
	ModelID string
}

func NewClient(ctx context.Context, region string, modelID string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("Unable to load AWS config: %w", err)
	}

	bedrockClient := bedrockruntime.NewFromConfig(cfg)

	return &Client{
		Client:  bedrockClient,
		ModelID: modelID,
	}, nil
}
