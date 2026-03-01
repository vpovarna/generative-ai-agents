package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/aggregator"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/api"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/config"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/executor"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/llm"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/models"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/prechecks"
	"github.com/rs/zerolog"
)

/*
INTEGRATION TEST CONCEPTS IN GO:

1. **What is an integration test?**
   - Tests multiple components working together (not just isolated units)
   - In this case: HTTP handler → executor → judges → aggregation
   - We test the REAL code path, not mocks

2. **How it differs from unit tests:**
   - Unit test: Test ONE function with mocks
   - Integration test: Test MULTIPLE components together

3. **Key Go patterns:**
   - httptest.NewRecorder() - Captures HTTP responses
   - httptest.NewRequest() - Creates fake HTTP requests
   - restful.Container - Real REST framework (not mocked)

4. **What we're testing:**
   - Real HTTP requests → Real API handlers → Real executors
   - We use MockLLMClient (only LLM calls are mocked, everything else is real)
*/

// setupTestAPI creates a complete API with real components but mocked LLM
func setupTestAPI(t *testing.T) (*restful.Container, *zerolog.Logger) {
	// 0. Set config path for tests (since tests run from internal/api/)
	// Tests need to find configs/judges.yaml from the project root
	os.Setenv("JUDGES_CONFIG_PATH", "../../configs/judges.yaml")

	// 1. Logger (silent for tests)
	logger := zerolog.Nop()

	// 2. Mock LLM Client (only external dependency we mock)
	// Why mock? Because we don't want to call real AWS Bedrock in tests
	mockLLM := &MockLLMClient{}

	// 3. PreChecks (REAL implementation)
	stageRunner := prechecks.NewStageRunner([]prechecks.Checker{
		&prechecks.LengthChecker{},
		&prechecks.OverlapChecker{MinOverlapThreshold: 0.3},
		&prechecks.FormatChecker{},
	})

	// 4. Judges (REAL implementation, but with mocked LLM)
	judgesConfig, err := config.LoadJudgesConfig()
	if err != nil {
		t.Fatalf("Failed to load judges config: %v", err)
	}

	judgePool := judge.NewJudgePool(mockLLM, &logger)
	judges, err := judgePool.BuildFromConfig(judgesConfig)
	if err != nil {
		t.Fatalf("Failed to build judges: %v", err)
	}

	judgeRunner := judge.NewJudgeRunner(judges, &logger)
	judgeFactory := judge.NewJudgeFactory(judges, &logger)

	// 5. Aggregator (REAL implementation)
	agg := aggregator.NewAggregator(aggregator.Weights{
		PreChecks: 0.3,
		LLMJudge:  0.7,
	}, &logger)

	// 6. Executors (REAL implementation)
	exec := executor.NewExecutor(stageRunner, judgeRunner, agg, 0.2, &logger)
	judgeExec := executor.NewJudgeExecutor(judgeFactory, &logger)

	// 7. API Handler (REAL implementation)
	handler := api.NewHandler(exec, judgeExec, &logger)

	// 8. REST Container (REAL framework)
	container := restful.NewContainer()
	api.RegisterRoutes(container, handler)

	return container, &logger
}

/*
TEST 1: Health Check
Purpose: Verify the API is running and responds to health checks
*/
func TestAPI_Health(t *testing.T) {
	// Setup: Create real API
	container, _ := setupTestAPI(t)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)

	// Create recorder to capture response
	recorder := httptest.NewRecorder()

	// Execute: Send request through real API
	container.ServeHTTP(recorder, req)

	// Assert: Check response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var response api.HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	t.Logf("✓ Health check passed: %+v", response)
}

/*
TEST 2: Full Evaluation - Happy Path
Purpose: Test complete evaluation pipeline with all judges
*/
func TestAPI_Evaluate_FullPipeline(t *testing.T) {
	// Setup
	container, _ := setupTestAPI(t)

	// Create evaluation request (happy case: good answer)
	evalRequest := models.EvaluationRequest{
		EventID: "test-001",
		Interaction: models.Interaction{
			UserQuery: "What is the capital of France?",
			Answer:    "The capital of France is Paris.",
			Context:   "France is a country in Europe. Paris is its capital city.",
		},
	}

	// Marshal to JSON
	body, err := json.Marshal(evalRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create recorder
	recorder := httptest.NewRecorder()

	// Execute
	container.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
	}

	var result models.EvaluationResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify result structure
	if result.ID != "test-001" {
		t.Errorf("Expected ID 'test-001', got '%s'", result.ID)
	}

	if len(result.Stages) == 0 {
		t.Error("Expected stages in result, got none")
	}

	// Verify both prechecks and judges ran
	hasPrechecks := false
	hasJudges := false
	for _, stage := range result.Stages {
		if stage.Name == "length-checker" || stage.Name == "overlap-checker" || stage.Name == "format-checker" {
			hasPrechecks = true
		}
		if stage.Name == "relevance-judge" || stage.Name == "coherence-judge" {
			hasJudges = true
		}
	}

	if !hasPrechecks {
		t.Error("Expected precheck stages")
	}
	if !hasJudges {
		t.Error("Expected judge stages")
	}

	// Verify verdict is set
	if result.Verdict == "" {
		t.Error("Expected verdict to be set")
	}

	// Verify confidence is in valid range
	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Expected confidence in [0,1], got %f", result.Confidence)
	}

	t.Logf("✓ Full evaluation passed:")
	t.Logf("  - ID: %s", result.ID)
	t.Logf("  - Verdict: %s", result.Verdict)
	t.Logf("  - Confidence: %.2f", result.Confidence)
	t.Logf("  - Stages: %d", len(result.Stages))
}

/*
TEST 3: Single Judge Evaluation
Purpose: Test evaluating with only one judge (faster endpoint)
*/
func TestAPI_EvaluateSingleJudge_Relevance(t *testing.T) {
	// Setup
	container, _ := setupTestAPI(t)

	// Create evaluation request
	evalRequest := models.EvaluationRequest{
		EventID: "test-002",
		Interaction: models.Interaction{
			UserQuery: "What is AI?",
			Answer:    "AI stands for Artificial Intelligence.",
		},
	}

	body, _ := json.Marshal(evalRequest)

	// Create HTTP request for relevance judge only
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/evaluate/judge/relevance?threshold=0.7",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Execute
	container.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
	}

	var result models.EvaluationResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify only relevance judge ran
	if len(result.Stages) != 1 {
		t.Errorf("Expected 1 stage (relevance-judge), got %d", len(result.Stages))
	}

	if len(result.Stages) > 0 && result.Stages[0].Name != "relevance-judge" {
		t.Errorf("Expected 'relevance-judge', got '%s'", result.Stages[0].Name)
	}

	t.Logf("✓ Single judge evaluation passed:")
	t.Logf("  - Judge: %s", result.Stages[0].Name)
	t.Logf("  - Score: %.2f", result.Stages[0].Score)
	t.Logf("  - Verdict: %s", result.Verdict)
}

/*
TEST 4: Faithfulness Judge (requires context)
Purpose: Test a judge that requires context field
*/
func TestAPI_EvaluateSingleJudge_Faithfulness(t *testing.T) {
	// Setup
	container, _ := setupTestAPI(t)

	// Create evaluation request WITH context
	evalRequest := models.EvaluationRequest{
		EventID: "test-003",
		Interaction: models.Interaction{
			UserQuery: "What does the documentation say about Redis?",
			Answer:    "Redis is used for streaming messages.",
			Context:   "The system uses Redis Streams for message queue functionality.", // Required!
		},
	}

	body, _ := json.Marshal(evalRequest)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/evaluate/judge/faithfulness",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Execute
	container.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
	}

	var result models.EvaluationResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(result.Stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(result.Stages))
	}

	if len(result.Stages) > 0 && result.Stages[0].Name != "faithfulness-judge" {
		t.Errorf("Expected 'faithfulness-judge', got '%s'", result.Stages[0].Name)
	}

	t.Logf("✓ Faithfulness judge passed:")
	t.Logf("  - Score: %.2f", result.Stages[0].Score)
	t.Logf("  - Reason: %s", result.Stages[0].Reason)
}

/*
TEST 5: Multiple Judges at Once
Purpose: Test evaluating with multiple specific judges
*/
func TestAPI_Evaluate_MultipleJudges(t *testing.T) {
	// Setup
	container, _ := setupTestAPI(t)

	// Test different judges with the same request
	judges := []string{"relevance", "coherence", "completeness", "instruction"}

	evalRequest := models.EvaluationRequest{
		EventID: "test-004",
		Interaction: models.Interaction{
			UserQuery: "Explain Go interfaces in one sentence.",
			Answer:    "Go interfaces define method signatures that types must implement.",
		},
	}

	for _, judgeName := range judges {
		body, _ := json.Marshal(evalRequest)

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/evaluate/judge/"+judgeName,
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		container.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("Judge %s failed with status %d", judgeName, recorder.Code)
			continue
		}

		var result models.EvaluationResult
		json.Unmarshal(recorder.Body.Bytes(), &result)

		t.Logf("✓ %s judge:", judgeName)
		t.Logf("  - Score: %.2f", result.Stages[0].Score)
		t.Logf("  - Verdict: %s", result.Verdict)
	}
}

/*
TEST 6: Early Exit Scenario
Purpose: Test that poor responses trigger early exit (skip LLM judges)
*/
func TestAPI_Evaluate_EarlyExit(t *testing.T) {
	// Setup
	container, _ := setupTestAPI(t)

	// Create request with very poor answer (should fail prechecks)
	evalRequest := models.EvaluationRequest{
		EventID: "test-005",
		Interaction: models.Interaction{
			UserQuery: "Explain quantum computing, its applications, and future implications?",
			Answer:    "Yes.", // Very short answer = early exit
		},
	}

	body, _ := json.Marshal(evalRequest)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Execute
	container.ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var result models.EvaluationResult
	json.Unmarshal(recorder.Body.Bytes(), &result)

	// Verify early exit: should have prechecks but NO judges
	hasPrechecks := false
	hasJudges := false
	for _, stage := range result.Stages {
		if stage.Name == "length-checker" {
			hasPrechecks = true
		}
		if stage.Name == "relevance-judge" {
			hasJudges = true
		}
	}

	if !hasPrechecks {
		t.Error("Expected prechecks to run")
	}

	if hasJudges {
		t.Error("Expected early exit - judges should NOT run for poor answer")
	}

	// Should be a fail verdict
	if result.Verdict != models.VerdictFail {
		t.Errorf("Expected 'fail' verdict for early exit, got '%s'", result.Verdict)
	}

	t.Logf("✓ Early exit test passed:")
	t.Logf("  - Verdict: %s (expected fail)", result.Verdict)
	t.Logf("  - Confidence: %.2f", result.Confidence)
	t.Logf("  - Stages: %d (prechecks only, no judges)", len(result.Stages))
}

/*
=============================================================================
MOCK LLM CLIENT
=============================================================================
This is the ONLY component we mock in integration tests.
Why? Because we don't want to make real API calls to AWS Bedrock during tests.

For integration tests, we mock ONLY external dependencies:
- ✅ Mock: LLM API calls (AWS Bedrock, OpenAI)
- ✅ Real: All internal business logic
- ✅ Real: HTTP handlers, executors, judges, aggregators
*/

type MockLLMClient struct{}

func (m *MockLLMClient) InvokeModel(ctx context.Context, request llm.LLMRequest) (*llm.LLMResponse, error) {
	// Return realistic mock response based on prompt content
	return &llm.LLMResponse{
		Content:    `{"score": 0.85, "reason": "Mock evaluation from test"}`,
		StopReason: "stop",
	}, nil
}

func (m *MockLLMClient) InvokeModelWithRetry(ctx context.Context, request llm.LLMRequest) (*llm.LLMResponse, error) {
	return m.InvokeModel(ctx, request)
}
