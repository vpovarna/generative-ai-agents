package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/api/middleware"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

type Handler struct {
	executor      *executor.Executor
	judgeExecutor *executor.JudgeExecutor
	logger        *zerolog.Logger
}

func NewHandler(executor *executor.Executor, judgeExecutor *executor.JudgeExecutor, logger *zerolog.Logger) *Handler {
	return &Handler{
		executor:      executor,
		judgeExecutor: judgeExecutor,
		logger:        logger,
	}
}

// POST /api/v1/evaluate
// Body: EvaluateRequest
// Returns: EvaluationResult
func (h *Handler) Evaluate(req *restful.Request, resp *restful.Response) {
	var evalRequest models.EvaluationRequest
	if err := req.ReadEntity(&evalRequest); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse request body")
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	h.logger.Info().
		Str("event_id", evalRequest.EventID).
		Str("event_type", string(evalRequest.EventType)).
		Str("agent_name", string(evalRequest.Agent.Name)).
		Msg("Start evaluation")

	ctx := req.Request.Context()
	evaluationContext := normalize(evalRequest)

	evalResult := h.executor.Execute(ctx, evaluationContext)

	h.logger.Info().
		Str("event_id", evalResult.ID).
		Str("verdict", string(evalResult.Verdict)).
		Float64("confidence", evalResult.Confidence).
		Msg("Evaluation complete")

	resp.WriteHeaderAndEntity(http.StatusOK, evalResult)
}

// POST /api/v1/evaluate/judge/{judge_name}
func (h *Handler) EvaluateSingleJudge(req *restful.Request, resp *restful.Response) {
	judgeName := req.PathParameter("judge_name")
	thresholdStr := req.QueryParameter("threshold")
	threshold := 0.7
	if thresholdStr != "" {
		if parsedThreshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			if parsedThreshold >= 0.0 && parsedThreshold <= 1.0 {
				threshold = parsedThreshold
			} else {
				h.logger.Warn().Str("threshold", thresholdStr).Msg("Invalid threshold, using default 0.7")
			}
		}
	}

	var evalRequest models.EvaluationRequest

	if err := req.ReadEntity(&evalRequest); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse request body")
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	h.logger.Info().
		Str("event_id", evalRequest.EventID).
		Str("judge_name", judgeName).
		Float64("threshold", threshold).
		Str("event_type", string(evalRequest.EventType)).
		Str("agent_name", string(evalRequest.Agent.Name)).
		Msg("Start evaluation")

	ctx := req.Request.Context()
	evalContext := normalize(evalRequest)

	evalResult := h.judgeExecutor.Execute(ctx, judgeName, threshold, evalContext)

	h.logger.Info().
		Str("judge_name", judgeName).
		Float64("threshold", threshold).
		Str("event_id", evalResult.ID).
		Str("verdict", string(evalResult.Verdict)).
		Float64("confidence", evalResult.Confidence).
		Msg("Evaluation complete")

	resp.WriteHeaderAndEntity(http.StatusOK, evalResult)

}

// Health handler GET API /api/v1/health
func (h *Handler) Health(req *restful.Request, resp *restful.Response) {
	healthResponse := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}

	resp.WriteHeaderAndEntity(http.StatusOK, healthResponse)
}

func normalize(req models.EvaluationRequest) models.EvaluationContext {
	return models.EvaluationContext{
		RequestID: req.EventID,
		Query:     req.Interaction.UserQuery,
		Context:   req.Interaction.Context,
		Answer:    req.Interaction.Answer,
		CreatedAt: time.Now(),
	}
}
