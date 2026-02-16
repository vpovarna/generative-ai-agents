package agent

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/agent"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/middleware"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/rewrite"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	bedrockClient *bedrock.Client
	rewriter      *rewrite.Rewriter
	modelID       string
}

func NewHandler(client *bedrock.Client, rewriter *rewrite.Rewriter, modelID string) *Handler {
	return &Handler{
		bedrockClient: client,
		rewriter:      rewriter,
		modelID:       modelID,
	}
}

// Query handles POST /api/v1/query
func (h *Handler) Query(req *restful.Request, resp *restful.Response) {
	var queryRequest agent.QueryRequest

	if err := req.ReadEntity(&queryRequest); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	queryRequest.SetDefaults()
	if err := queryRequest.Validate(); err != nil {
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("prompt", queryRequest.Prompt).
		Int("max_tokens", queryRequest.MaxToken).
		Float64("temperature", queryRequest.Temperature).
		Msg("Process Query")

	ctx := req.Request.Context()

	// Query rewrite
	rewrittenQuery, err := h.rewriter.RewriteQuery(ctx, queryRequest.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Query rewrite failed")
		// Continue with original query
		rewrittenQuery = queryRequest.Prompt
	}

	response, err := h.bedrockClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      rewrittenQuery,
		MaxTokens:   queryRequest.MaxToken,
		Temperature: queryRequest.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to invoke Claude")
		middleware.HandleError(resp, err, http.StatusInternalServerError)
		return
	}

	queryResponse := agent.QueryResponse{
		Content:    response.Content,
		StopReason: response.StopReason,
		Model:      h.modelID,
	}

	resp.WriteHeaderAndEntity(http.StatusOK, queryResponse)
}

// Query handles POST /api/v1/query/stream
func (h *Handler) QueryStream(req *restful.Request, resp *restful.Response) {
	var queryRequest agent.QueryRequest

	if err := req.ReadEntity(&queryRequest); err != nil {
		log.Error().Err(err).Msg("Unable to parse query request")
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	queryRequest.SetDefaults()
	if err := queryRequest.Validate(); err != nil {
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("prompt", queryRequest.Prompt).
		Int("max_tokens", queryRequest.MaxToken).
		Float64("temperature", queryRequest.Temperature).
		Msg("Process Query Stream")

	ctx := req.Request.Context()

	resp.AddHeader("Content-Type", "text/event-stream")
	resp.AddHeader("Cache-Control", "no-cache")
	resp.AddHeader("Connection", "keep-alive")
	resp.AddHeader("X-Accel-Buffering", "no")

	writer := resp.ResponseWriter
	flusher, ok := writer.(http.Flusher)
	if !ok {
		middleware.HandleError(resp, fmt.Errorf("streaming not supported"), http.StatusInternalServerError)
		return
	}

	// Query rewrite
	rewrittenQuery, err := h.rewriter.RewriteQuery(ctx, queryRequest.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Query rewrite failed")
		// Continue with original query
		rewrittenQuery = queryRequest.Prompt
	}

	// Send starting event
	startEvent := agent.SSEEvent{
		Event: "start",
		Data: agent.StreamStartEvent{
			Model: h.modelID,
		},
	}

	if formatEvent, err := startEvent.Format(); err == nil {
		fmt.Fprint(writer, formatEvent)
		flusher.Flush()
	}

	response, err := h.bedrockClient.InvokeModelStream(ctx, bedrock.ClaudeRequest{
		Prompt:      rewrittenQuery,
		MaxTokens:   queryRequest.MaxToken,
		Temperature: queryRequest.Temperature,
	}, func(chunk string) error {
		// Send chunk event
		chunkEvent := agent.SSEEvent{
			Event: "chunk",
			Data: agent.StreamChunkEvent{
				Text: chunk,
			},
		}

		if formatEvent, ok := chunkEvent.Format(); ok == nil {
			fmt.Fprint(writer, formatEvent)
			flusher.Flush()
		}

		return nil
	})

	if err != nil {
		// Send error event
		errorEvent := agent.SSEEvent{
			Event: "error",
			Data: agent.StreamErrorEvent{
				Error: err.Error(),
			},
		}

		if formatEvent, ok := errorEvent.Format(); ok == nil {
			fmt.Fprint(writer, formatEvent)
			flusher.Flush()
		}
		return
	}

	// Send end event
	doneEvent := agent.SSEEvent{
		Event: "done",
		Data: agent.StreamDoneEvent{
			StopReason: response.StopReason,
		},
	}
	if formatEvent, ok := doneEvent.Format(); ok == nil {
		fmt.Fprint(writer, formatEvent)
		flusher.Flush()
	}
}

// Health handler GET API /api/v1/health
func (h *Handler) Health(req *restful.Request, resp *restful.Response) {
	healthResponse := agent.HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}

	resp.WriteHeaderAndEntity(http.StatusOK, healthResponse)
}
