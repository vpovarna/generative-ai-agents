package agent

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/middleware"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Query handles POST /api/v1/query
func (h *Handler) Query(req *restful.Request, resp *restful.Response) {
	var queryRequest QueryRequest

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

	queryResponse, err := h.service.Query(ctx, queryRequest)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query")
		middleware.HandleError(resp, err, http.StatusInternalServerError)
		return
	}

	resp.WriteHeaderAndEntity(http.StatusOK, queryResponse)
}

// Query handles POST /api/v1/query/stream
func (h *Handler) QueryStream(req *restful.Request, resp *restful.Response) {
	var queryRequest QueryRequest

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

	err := h.service.QueryStream(ctx, queryRequest, flusher, writer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query stream")
		middleware.HandleError(resp, err, http.StatusInternalServerError)
		return
	}

	flusher.Flush()
}

// Health handler GET API /api/v1/health
func (h *Handler) Health(req *restful.Request, resp *restful.Response) {
	healthResponse := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}

	resp.WriteHeaderAndEntity(http.StatusOK, healthResponse)
}
