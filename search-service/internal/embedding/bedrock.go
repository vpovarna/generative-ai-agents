package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/rs/zerolog/log"
)

var titanEmbeddingModelID string = "amazon.titan-embed-text-v2:0"

type BedrockEmbedder struct {
	client  *bedrockruntime.Client
	modelID string
}

type BedrockEmbedderRequest struct {
	InputText string `json:"inputText"`
}

type BedrockEmbedderResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

func NewBedrockEmbedder(bedrockClient *bedrockruntime.Client) *BedrockEmbedder {
	return &BedrockEmbedder{
		client:  bedrockClient,
		modelID: titanEmbeddingModelID,
	}
}

func (e *BedrockEmbedder) GenerateEmbeddings(ctx context.Context, query string) ([]float32, error) {

	payload := BedrockEmbedderRequest{
		InputText: query,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Unable to serialize user query")
		return nil, err
	}

	response, err := e.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		Body:        bytes,
		ModelId:     aws.String(e.modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})

	if err != nil {
		log.Error().Err(err).Msg("Unable to query embedding model")
		return nil, err
	}

	var bedrockEmbedderResponse BedrockEmbedderResponse

	err = json.Unmarshal(response.Body, &bedrockEmbedderResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed embedder response")
		return nil, err
	}

	return bedrockEmbedderResponse.Embedding, nil
}

func (e *BedrockEmbedder) GenerateBatchEmbeddings(ctx context.Context, queries []string) ([][]float32, error) {
	const CONCURRENT_LIMIT = 5 // Don't overwhelm API

	embeddings := make([][]float32, len(queries))
	errors := make([]error, len(queries))

	// Process in groups of 5 concurrent requests
	for i := 0; i < len(queries); i += CONCURRENT_LIMIT {
		end := i + CONCURRENT_LIMIT
		if end > len(queries) {
			end = len(queries)
		}

		// Use sync.WaitGroup for this batch
		var wg sync.WaitGroup
		for j := i; j < end; j++ {
			wg.Add(1)
			go func(index int, query string) {
				defer wg.Done()
				embedding, err := e.GenerateEmbeddings(ctx, query)
				if err != nil {
					errors[index] = err
					return
				}
				embeddings[index] = embedding
			}(j, queries[j])
		}
		wg.Wait()

		// Check for errors in this batch
		for j := i; j < end; j++ {
			if errors[j] != nil {
				return nil, fmt.Errorf("failed to generate embedding for query %d: %w", j, errors[j])
			}
		}
	}

	return embeddings, nil
}
