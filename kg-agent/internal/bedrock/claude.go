package bedrock

import (
	"strings"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// ClaudeRequest is the message sent to Claude
type ClaudeRequest struct {
	Prompt      string
	MaxTokens   int
	Temperature float64
}

// ClaudeResponse is the Claude's response
type ClaudeResponse struct {
	Content    string
	StopReason string
}

// Claude API request format (what Bedrock expects)
type claudeMessageRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxToken         int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature,omitempty"`
	Messages         []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Claude API response format (what Bedrock returns)
type claudeMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

var anthropic_version = "bedrock-2023-05-31"

func (c *Client) InvokeModel(ctx context.Context, request ClaudeRequest) (*ClaudeResponse, error) {
	payload := claudeMessageRequest{
		AnthropicVersion: anthropic_version,
		MaxToken:         request.MaxTokens,
		Temperature:      request.Temperature,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: request.Prompt,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %w", request)
	}

	// Call Bedrock
	output, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &c.modelID,
		Body:        body,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke mode: %w", err)
	}

	// Parse response
	var response claudeMessageResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bedrock response: %w", err)
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

type StreamCallback func(chunk string) error

func (c *Client) InvokeModelStream(ctx context.Context, req ClaudeRequest, callback StreamCallback) (*ClaudeResponse, error) {
	payload := claudeMessageRequest{
		AnthropicVersion: anthropic_version,
		MaxToken:         req.MaxTokens,
		Temperature:      req.Temperature,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
	}

	// Marshall to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %w", err)
	}

	output, err := c.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     &c.modelID,
		Body:        body,
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to invoke mode stream: %w", err)
	}

	// TODO: Test this with channel
	// Process the stream
	stream := output.GetStream()
	defer stream.Close()

	var fullContent strings.Builder
	var stopReason string

	// Read events from the stream
	for event := range stream.Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			// Parse the chunk - Claude sends different event types
			var chunkResponse struct {
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
				ContentBlock struct {
					Text string `json:"text"`
				} `json:"content_block"`
				Message struct {
					StopReason string `json:"stop_reason"`
				} `json:"message"`
			}

			if err := json.Unmarshal(v.Value.Bytes, &chunkResponse); err != nil {
				// Just skip chunks we can't parse
				continue
			}

			// Extract text from delta (streaming text chunks)
			if chunkResponse.Delta.Text != "" {
				fullContent .WriteString(chunkResponse.Delta.Text)
				if callback != nil {
					if err := callback(chunkResponse.Delta.Text); err != nil {
						return nil, fmt.Errorf("callback error: %w", err)
					}
				}
			}

			// Extract text from content_block (initial content)
			if chunkResponse.ContentBlock.Text != "" {
				fullContent .WriteString(chunkResponse.ContentBlock.Text)
				if callback != nil {
					if err := callback(chunkResponse.ContentBlock.Text); err != nil {
						return nil, fmt.Errorf("callback error: %w", err)
					}
				}
			}

			// Capture stop reason if present
			if chunkResponse.Message.StopReason != "" {
				stopReason = chunkResponse.Message.StopReason
			}

		default:
			// Ignore other event types we don't need
			continue
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	return &ClaudeResponse{
		Content:    fullContent.String(),
		StopReason: stopReason,
	}, nil
}
