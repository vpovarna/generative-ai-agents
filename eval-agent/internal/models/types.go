package models

import (
	"time"
)

type Verdict string

const (
	VerdictPass   Verdict = "pass"
	VerdictFail   Verdict = "fail"
	VerdictReview Verdict = "review"
)

type EventType string

const (
	EventTypeAgentResponse EventType = "agent_response"
	EventTypeAgentError    EventType = "agent_error"
)

type Agent struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type Interaction struct {
	UserQuery string `json:"user_query"`
	Context   string `json:"context"`
	Answer    string `json:"answer"`
}

// Input message

type EvaluationRequest struct {
	EventID     string      `json:"event_id"`
	EventType   EventType   `json:"event_type"`
	Agent       Agent       `json:"agent"`
	Interaction Interaction `json:"interaction"`
}

// Normalized internal object
type EvaluationContext struct {
	RequestID string    `json:"request_id" jsonschema:"required,description=Unique event identifier"`
	Query     string    `json:"user_query" jsonschema:"required,description=User's original query"`
	Context   string    `json:"context,omitempty" jsonschema:"description=Optional context or retrieved documents"`
	Answer    string    `json:"answer" jsonschema:"required,description=Agent response to evaluate"`
	CreatedAt time.Time `json:"created_at" jsonschema:"description=Time when the evaluation context was created"`
}

// One evaluator's output
type StageResult struct {
	Name     string        `json:"name"`
	Score    float64       `json:"score"`
	Reason   string        `json:"reason"`
	Duration time.Duration `json:"duration_ns"`
}

// Final output emitted to Kafka
type EvaluationResult struct {
	ID         string        `json:"id"`
	Stages     []StageResult `json:"stages"`
	Confidence float64       `json:"confidence"`
	Verdict    Verdict       `json:"verdict"`
}
