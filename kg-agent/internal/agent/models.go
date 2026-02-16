package agent

import (
	"encoding/json"
	"fmt"

	"github.com/povarna/generative-ai-with-go/kg-agent/internal/middleware"
)

type QueryRequest struct {
	Prompt      string  `json:"prompt" description:"The prompt to send to Claude"`
	MaxToken    int     `json:"max_tokens,omitempty" description:"Maximum Tokens to generate (default: 2000)"`
	Temperature float64 `json:"temperature,omitempty" description:"Temperature for generation (0.0-1.0, default:0.0)"`
}

type QueryResponse struct {
	Content    string `json:"content" description:"Claude's response text"`
	StopReason string `json:"stop_reason" description:"Why generation stopped"`
	Model      string `json:"model" description:"Model ID used"`
}

type HealthResponse struct {
	Status  string `json:"status" description:"Service status"`
	Version string `json:"version" description:"API version"`
}

type ErrorResponse struct {
	Error   string `json:"error" description:"Error message"`
	Code    int    `json:"code" description:"HTTP status code"`
	Details string `json:"details" description:"Additional error details"`
}

func (q *QueryRequest) Validate() error {
	if q.Prompt == "" {
		return middleware.ErrEmptyPrompt
	}

	if q.MaxToken < 0 || q.MaxToken > 100000 {
		return middleware.ErrInvalidMaxTokens
	}

	if q.Temperature < 0.0 || q.Temperature > 1.0 {
		return middleware.ErrInvalidTemperature
	}
	return nil
}

func (q *QueryRequest) SetDefaults() {
	if q.MaxToken == 0 {
		q.MaxToken = 2000
	}

	if q.Temperature == 0 {
		q.Temperature = 0.0
	}
}

type SSEEvent struct {
	Event string      `json:"-"`
	Data  interface{} `json:"-"`
}

// SEE Event data structures
type StreamStartEvent struct {
	Model string `json:"model"`
}

type StreamChunkEvent struct {
	Text string `json:"text"`
}

type StreamDoneEvent struct {
	StopReason string `json:"stop_reason"`
}

type StreamErrorEvent struct {
	Error string `json:"error"`
}

func (e SSEEvent) Format() (string, error) {
	jsonData, err := json.Marshal(e.Data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("event: %s\ndata: %s\n\n", e.Event, string(jsonData)), nil
}
