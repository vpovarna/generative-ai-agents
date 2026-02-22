package bedrock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type ClaudeRequest struct {
	Pormpt      string
	MaxTokens   int
	Temperature float64
}

type ClaudeResponse struct {
	Content    string
	StopReason string
}

type claudeMessageRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature"`
	Messages         []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

var anthropicVersion = "bedrock-2023-05-31"

func (c *Client) InvokeModel(ctx context.Context, request ClaudeRequest) (*ClaudeResponse, error) {
	payload := claudeMessageRequest{
		AnthropicVersion: anthropicVersion,
		MaxTokens:        request.MaxTokens,
		Temperature:      request.Temperature,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: request.Pormpt,
			},
		},
	}

	byes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Unable to serialize claude request. Error: %w", err)
	}

	output, err := c.Client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &c.ModelID,
		Body:        byes,
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to invoke claude mode. Error: %w", err)
	}

	var response claudeMessageResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal bedrock response. Error: %w", err)
	}

	// Extract the response
	var content string
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}

	return &ClaudeResponse{
		Content:    content,
		StopReason: response.StopReason,
	}, nil
}
