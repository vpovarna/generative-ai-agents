# API Test Cases

Comprehensive test scenarios for the eval-agent HTTP API.

## Setup

### Prerequisites

Ensure your `.env` file is configured:

```env
# AWS Bedrock credentials (if using anthropic models)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret

# OpenAI credentials (if using openai models)
OPEN_AI_KEY=sk-...

# API configuration
EVAL_AGENT_API_PORT=18082
EARLY_EXIT_THRESHOLD=0.2
```

Judges are configured in `configs/judges.yaml`. The system dynamically creates LLM clients based on models referenced in the configuration.

### Start the Server

```bash
cd eval-agent
go run cmd/api/main.go
```

Server runs on `http://localhost:18082`

**Expected startup logs:**
```
INFO judge created successfully judge=relevance model_family=anthropic model_id=us.anthropic.claude-3-5-haiku-20241022-v1:0
INFO judge created successfully judge=faithfulness model_family=anthropic model_id=us.anthropic.claude-3-5-haiku-20241022-v1:0
INFO judge created successfully judge=coherence model_family=anthropic model_id=us.anthropic.claude-3-5-haiku-20241022-v1:0
INFO judge created successfully judge=completeness model_family=anthropic model_id=us.anthropic.claude-3-5-haiku-20241022-v1:0
INFO judge created successfully judge=instruction model_family=anthropic model_id=us.anthropic.claude-3-5-haiku-20241022-v1:0
INFO judge pool built successfully total_judges=5
```

## Health Check Tests

### Test Case 1: Health Endpoint

**Request:**
```bash
curl http://localhost:18082/api/v1/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

**Status Code:** 200

---

## Full Pipeline Evaluation Tests

### Test Case 2: Happy Path - High Quality Answer

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-001",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe. Paris is its capital city and largest metropolis.",
      "answer": "The capital of France is Paris."
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-001",
  "stages": [
    {"name": "length-checker", "score": 1.0, "reason": "...", "duration_ns": 15000},
    {"name": "overlap-checker", "score": 0.85, "reason": "...", "duration_ns": 12000},
    {"name": "format-checker", "score": 1.0, "reason": "...", "duration_ns": 10000},
    {"name": "relevance-judge", "score": 0.95, "reason": "...", "duration_ns": 850000000},
    {"name": "faithfulness-judge", "score": 1.0, "reason": "...", "duration_ns": 820000000}
  ],
  "confidence": 0.92,
  "verdict": "pass"
}
```

**Expected:**
- Status Code: 200
- `confidence` > 0.8
- `verdict` = "pass"
- 8 stages (3 prechecks + 5 judges)
- Response time: ~3-4 seconds (judges run in parallel)

### Test Case 3: Early Exit - Very Short Answer

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-002",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Explain quantum computing in detail",
      "context": "Quantum computing uses quantum mechanics principles...",
      "answer": "Yes."
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-002",
  "stages": [
    {"name": "length-checker", "score": 0.0, "reason": "Answer too short", "duration_ns": 12000},
    {"name": "overlap-checker", "score": 0.0, "reason": "...", "duration_ns": 10000},
    {"name": "format-checker", "score": 0.5, "reason": "...", "duration_ns": 8000}
  ],
  "confidence": 0.15,
  "verdict": "fail"
}
```

**Expected:**
- Status Code: 200
- `confidence` < 0.2
- `verdict` = "fail"
- Only 3 stages (prechecks only, early exit triggered)
- No LLM judges called (cost savings)

### Test Case 4: Fail Verdict - Irrelevant Answer

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-003",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What are the main causes of climate change?",
      "context": "Climate change is primarily caused by greenhouse gas emissions from human activities.",
      "answer": "There are various factors that contribute to weather patterns."
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-003",
  "stages": [
    {"name": "length-checker", "score": 1.0, "duration_ns": 15000},
    {"name": "overlap-checker", "score": 0.5, "duration_ns": 12000},
    {"name": "format-checker", "score": 1.0, "duration_ns": 10000},
    {"name": "relevance-judge", "score": 0.2, "duration_ns": 2388000000},
    {"name": "faithfulness-judge", "score": 0.2, "duration_ns": 2372000000},
    {"name": "coherence-judge", "score": 0.3, "duration_ns": 2994000000},
    {"name": "completeness-judge", "score": 0.0, "duration_ns": 2081000000},
    {"name": "instruction-judge", "score": 0.4, "duration_ns": 1806000000}
  ],
  "confidence": 0.354,
  "verdict": "fail"
}
```

**Expected:**
- Status Code: 200
- `confidence` ~0.35
- `verdict` = "fail"
- 8 stages (full pipeline - all judges run)
- Low scores across all judges:
  - **relevance: 0.2** - Answer discusses weather patterns, not climate change causes
  - **faithfulness: 0.2** - Doesn't mention greenhouse gases from context
  - **coherence: 0.3** - Vague but not contradictory
  - **completeness: 0.0** - Completely fails to address the query
  - **instruction: 0.4** - No specific instructions violated
- Total duration: ~3 seconds (5 parallel LLM calls)

### Test Case 5: Hallucination Detection

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-004",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the population of Tokyo?",
      "context": "Tokyo is the capital of Japan.",
      "answer": "Tokyo has a population of 50 million people and is the largest city in China."
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-004",
  "stages": [
    {"name": "format-checker", "score": 1.0, "duration_ns": 8292},
    {"name": "length-checker", "score": 1.0, "duration_ns": 83},
    {"name": "overlap-checker", "score": 0.67, "duration_ns": 7833},
    {
      "name": "faithfulness-judge",
      "score": 0.1,
      "reason": "Answer contains multiple factual errors not present in the original context: incorrectly states population, incorrectly identifies country, and introduces unsupported claims about city size",
      "duration_ns": 2142318792
    },
    {
      "name": "coherence-judge",
      "score": 0.2,
      "reason": "Multiple logical inconsistencies: Tokyo is in Japan, not China; population figure is inaccurate; statements contradict known geographic and demographic facts",
      "duration_ns": 1780915375
    },
    {
      "name": "relevance-judge",
      "score": 0.2,
      "reason": "The answer contains a factual error about Tokyo's population and location. While it mentions a population number, the details are incorrect. Tokyo is in Japan, not China...",
      "duration_ns": 2326646583
    },
    {
      "name": "completeness-judge",
      "score": 0.0,
      "reason": "Incorrect population figure, incorrect country, factually inaccurate answer",
      "duration_ns": 1474770625
    },
    {
      "name": "instruction-judge",
      "score": 0.4,
      "reason": "Factual information is incorrect (Tokyo is in Japan, not China) and population figure is inaccurate...",
      "duration_ns": 2206704875
    }
  ],
  "confidence": 0.393,
  "verdict": "fail"
}
```

**Expected:**
- Status Code: 200
- `confidence` = 0.393
- `verdict` = "fail"
- 8 stages (full pipeline)
- **Excellent hallucination detection** - All judges identify the errors:
  - **faithfulness: 0.1** - Correctly flags factual errors not in context ("China" not mentioned)
  - **coherence: 0.2** - Identifies logical inconsistency (Tokyo cannot be in both Japan and China)
  - **relevance: 0.2** - Notes answer addresses population but with wrong facts
  - **completeness: 0.0** - Answer is factually wrong
  - **instruction: 0.4** - No specific format instructions, penalizes factual errors
- Total duration: ~2.3 seconds (judges run in parallel)
- System successfully detects:
  - Geographic hallucination (China vs Japan)
  - Population exaggeration (50M vs actual ~14M city, ~37M metro)
  - Contradictory claims

---

## Single Judge Evaluation Tests

### Test Case 6: Relevance Judge

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/relevance \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-005",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is machine learning?",
      "context": "ML is a subset of AI.",
      "answer": "Machine learning is a method where computers learn from data without explicit programming."
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-005",
  "stages": [
    {"name": "relevance-judge", "score": 0.95, "reason": "Answer directly addresses the query", "duration_ns": 850000000}
  ],
  "confidence": 0.95,
  "verdict": "pass"
}
```

**Expected:**
- Status Code: 200
- Only 1 stage (relevance-judge)
- Score close to 1.0 for relevant answer

### Test Case 7: Custom Threshold

**Request:**
```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/faithfulness?threshold=0.9" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-006",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the boiling point of water?",
      "context": "Water boils at 100°C at sea level.",
      "answer": "Water boils at 100 degrees Celsius."
    }
  }'
```

**Expected Response:**
- Status Code: 200
- High faithfulness score (grounded in context)
- `verdict` = "pass" (score > 0.9 threshold)

### Test Case 8: Invalid Judge Name

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/invalid-judge \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-007",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Test",
      "context": "Test",
      "answer": "Test"
    }
  }'
```

**Expected Response:**
- Status Code: 400 or 404
- Error message: "judge not found" or similar

---

## Error Handling Tests

### Test Case 9: Missing Required Fields

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-008",
    "interaction": {
      "user_query": "Test"
    }
  }'
```

**Expected Response:**
- Status Code: 400
- Error message about missing required fields

### Test Case 10: Invalid JSON

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
```

**Expected Response:**
- Status Code: 400
- Error message: "Failed to parse request body" or similar

### Test Case 11: Empty Answer

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-009",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is AI?",
      "context": "AI is artificial intelligence.",
      "answer": ""
    }
  }'
```

**Expected Response:**
- Status Code: 200
- Very low precheck scores
- Early exit with `verdict` = "fail"

---

## Performance Tests

### Test Case 12: Concurrent Requests

**Script:**
```bash
# Send 10 concurrent requests
for i in {1..10}; do
  curl -X POST http://localhost:18082/api/v1/evaluate \
    -H "Content-Type: application/json" \
    -d "{\"event_id\":\"perf-$i\",\"event_type\":\"agent_response\",\"agent\":{\"name\":\"test\",\"type\":\"rag\",\"version\":\"1.0\"},\"interaction\":{\"user_query\":\"Test\",\"context\":\"Test\",\"answer\":\"Test\"}}" &
done
wait
```

**Expected:**
- All 10 requests complete successfully
- Response times < 5 seconds per request
- No race conditions or errors

### Test Case 13: Large Context (10KB)

**Request:**
```bash
# Generate large context
LARGE_CONTEXT=$(python3 -c "print('Context word. ' * 2000)")

curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d "{
    \"event_id\": \"test-010\",
    \"event_type\": \"agent_response\",
    \"agent\": {\"name\": \"test\", \"type\": \"rag\", \"version\": \"1.0\"},
    \"interaction\": {
      \"user_query\": \"Summarize the context\",
      \"context\": \"$LARGE_CONTEXT\",
      \"answer\": \"This is a summary.\"
    }
  }"
```

**Expected:**
- Status Code: 200
- Evaluation completes without timeout
- All judges handle large context

---

## Edge Cases

### Test Case 14: Special Characters in Answer

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-011",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Show me code",
      "context": "Python example",
      "answer": "def hello():\n    print(\"Hello, World!\")\n    return True"
    }
  }'
```

**Expected:**
- Status Code: 200
- Evaluation handles newlines and special characters
- Format checker passes

### Test Case 15: Non-English Text

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-012",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Quelle est la capitale de la France?",
      "context": "La France est un pays en Europe.",
      "answer": "La capitale de la France est Paris."
    }
  }'
```

**Expected:**
- Status Code: 200
- Evaluation works with non-English text
- Judges provide appropriate scores

---

## Multi-Provider Testing

### Test Case 16: Mixed Provider Evaluation

**Setup:**
Update `configs/judges.yaml` to use different providers:

```yaml
judges:
  default_model:
    modelFamily: "anthropic"
    modelID: us.anthropic.claude-3-5-haiku-20241022-v1:0

  evaluators:
    - name: relevance
      model:
        modelFamily: "anthropic"
        modelID: us.anthropic.claude-3-5-haiku-20241022-v1:0
    - name: faithfulness
      model:
        modelFamily: "openai"
        modelID: gpt-4o-mini
```

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-multi-provider",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is machine learning?",
      "context": "ML is a subset of AI that enables systems to learn from data.",
      "answer": "Machine learning allows computers to learn patterns from data without explicit programming."
    }
  }'
```

**Expected:**
- Status Code: 200
- Startup logs show both Anthropic and OpenAI clients initialized
- Each judge uses its configured provider
- All judges complete successfully
- Mixed provider evaluation works seamlessly

---

## Correctness Judge Tests (Ground Truth Comparison)

The correctness judge evaluates semantic similarity between an answer and expected output (ground truth). It's disabled by default and requires `expected_output` field in requests.

### Test Case 17: Enable Correctness Judge

**Setup:**
Enable the correctness judge in `configs/judges.yaml`:

```yaml
judges:
  evaluators:
    # ... other judges ...

    - name: correctness
      enabled: true  # Change from false to true
      description: "Evaluates semantic similarity between answer and expected output"
      requires_context: false
      requires_expected_output: true
      prompt: |
        You are a correctness evaluation judge.
        Compare the provided answer with the expected output (ground truth).
        Score based on semantic equivalence, not exact string match.

        Answer: {{.Answer}}
        Expected Output: {{.ExpectedOutput}}

        Scoring guidelines:
        - 1.0: Semantically identical
        - 0.8-0.9: Mostly correct, minor differences
        - 0.5-0.7: Partially correct
        - 0.2-0.4: Somewhat related but different
        - 0.0-0.1: Completely different

        Respond ONLY in raw JSON with no markdown, no code blocks, no explanation:
        {"score": <float>, "reason": "<string>"}
      model:
        max_tokens: 200
        temperature: 0.0
        retry: true
```

**Restart the API server** after enabling the judge.

### Test Case 18: Single Correctness Judge - Exact Match

**Request:**
```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/correctness?threshold=0.8" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-correctness-001",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "qa", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of France?",
      "answer": "Paris",
      "expected_output": "Paris"
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-correctness-001",
  "stages": [
    {
      "name": "correctness-judge",
      "score": 1.0,
      "reason": "The answer and expected output are semantically identical",
      "duration_ns": 1234567890
    }
  ],
  "confidence": 1.0,
  "verdict": "pass"
}
```

**Expected:**
- Status Code: 200
- `confidence` = 1.0 (exact match)
- `verdict` = "pass"
- Only 1 stage (correctness judge)
- Response time: ~1-2 seconds

### Test Case 19: Single Correctness Judge - Semantic Match

**Request:**
```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/correctness?threshold=0.8" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-correctness-002",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "qa", "version": "1.0"},
    "interaction": {
      "user_query": "What is 2+2?",
      "answer": "The answer is four",
      "expected_output": "4"
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-correctness-002",
  "stages": [
    {
      "name": "correctness-judge",
      "score": 0.9,
      "reason": "Semantically equivalent - both convey the same answer (four/4), just different representations",
      "duration_ns": 1456789012
    }
  ],
  "confidence": 0.9,
  "verdict": "pass"
}
```

**Expected:**
- Status Code: 200
- `confidence` ~0.9 (semantic match despite different format)
- `verdict` = "pass"
- Demonstrates that judge understands "four" = "4"

### Test Case 20: Single Correctness Judge - Wrong Answer

**Request:**
```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/correctness?threshold=0.7" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-correctness-003",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "qa", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of Italy?",
      "answer": "Milan",
      "expected_output": "Rome"
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-correctness-003",
  "stages": [
    {
      "name": "correctness-judge",
      "score": 0.1,
      "reason": "Completely different answers - Milan is a city in Italy but not the capital. Expected Rome.",
      "duration_ns": 1345678901
    }
  ],
  "confidence": 0.1,
  "verdict": "fail"
}
```

**Expected:**
- Status Code: 200
- `confidence` ~0.1 (wrong answer)
- `verdict` = "fail"
- Correctly identifies that Milan ≠ Rome

### Test Case 21: Full Pipeline with Correctness - All Judges Pass

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-correctness-pipeline-001",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe. Paris is its capital city.",
      "answer": "The capital of France is Paris.",
      "expected_output": "Paris"
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-correctness-pipeline-001",
  "stages": [
    {"name": "length-checker", "score": 1.0, "duration_ns": 15000},
    {"name": "overlap-checker", "score": 0.85, "duration_ns": 12000},
    {"name": "format-checker", "score": 1.0, "duration_ns": 10000},
    {"name": "relevance-judge", "score": 0.95, "duration_ns": 1850000000},
    {"name": "faithfulness-judge", "score": 1.0, "duration_ns": 1820000000},
    {"name": "coherence-judge", "score": 1.0, "duration_ns": 1780000000},
    {"name": "completeness-judge", "score": 1.0, "duration_ns": 1750000000},
    {"name": "instruction-judge", "score": 1.0, "duration_ns": 1690000000},
    {"name": "correctness-judge", "score": 1.0, "reason": "Semantically identical to expected output", "duration_ns": 1650000000}
  ],
  "confidence": 0.96,
  "verdict": "pass"
}
```

**Expected:**
- Status Code: 200
- `confidence` > 0.9
- `verdict` = "pass"
- **9 stages** (3 prechecks + 6 judges including correctness)
- All judges score high
- Response time: ~3-4 seconds (judges run in parallel)

### Test Case 22: Full Pipeline with Correctness - Wrong Answer

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-correctness-pipeline-002",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is 10 + 5?",
      "context": "Basic arithmetic operations.",
      "answer": "The answer is 20.",
      "expected_output": "15"
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-correctness-pipeline-002",
  "stages": [
    {"name": "length-checker", "score": 1.0, "duration_ns": 14000},
    {"name": "overlap-checker", "score": 0.6, "duration_ns": 11000},
    {"name": "format-checker", "score": 1.0, "duration_ns": 9000},
    {"name": "relevance-judge", "score": 0.8, "duration_ns": 1950000000},
    {"name": "faithfulness-judge", "score": 0.9, "duration_ns": 1880000000},
    {"name": "coherence-judge", "score": 1.0, "duration_ns": 1820000000},
    {"name": "completeness-judge", "score": 0.9, "duration_ns": 1790000000},
    {"name": "instruction-judge", "score": 1.0, "duration_ns": 1750000000},
    {"name": "correctness-judge", "score": 0.0, "reason": "Completely different - answer says 20 but expected output is 15", "duration_ns": 1680000000}
  ],
  "confidence": 0.65,
  "verdict": "review"
}
```

**Expected:**
- Status Code: 200
- `confidence` ~0.65 (dragged down by correctness score of 0.0)
- `verdict` = "review" or "fail"
- 9 stages (full pipeline)
- Other judges score reasonably (answer is coherent, relevant, faithful to context)
- **Correctness judge: 0.0** - correctly identifies 20 ≠ 15
- Demonstrates that correctness judge catches factual errors even when other quality dimensions are fine

### Test Case 23: Correctness Judge Auto-Skip (No Expected Output)

**Request:**
```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-correctness-skip",
    "event_type": "agent_response",
    "agent": {"name": "test-agent", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is machine learning?",
      "context": "ML is a subset of AI.",
      "answer": "Machine learning is a method where computers learn from data."
    }
  }'
```

**Expected Response:**
```json
{
  "id": "test-correctness-skip",
  "stages": [
    {"name": "length-checker", "score": 1.0, "duration_ns": 15000},
    {"name": "overlap-checker", "score": 0.75, "duration_ns": 12000},
    {"name": "format-checker", "score": 1.0, "duration_ns": 10000},
    {"name": "relevance-judge", "score": 0.95, "duration_ns": 1850000000},
    {"name": "faithfulness-judge", "score": 0.9, "duration_ns": 1820000000},
    {"name": "coherence-judge", "score": 1.0, "duration_ns": 1780000000},
    {"name": "completeness-judge", "score": 0.9, "duration_ns": 1750000000},
    {"name": "instruction-judge", "score": 1.0, "duration_ns": 1690000000}
  ],
  "confidence": 0.89,
  "verdict": "pass"
}
```

**Expected:**
- Status Code: 200
- **8 stages** (3 prechecks + 5 judges - correctness judge auto-skipped)
- No correctness stage in results (because `expected_output` not provided)
- No errors or warnings
- Demonstrates backwards compatibility - requests without `expected_output` work unchanged

**Server logs:**
```
DEBUG skipping judge - expected_output not provided judge=correctness
DEBUG all judges completed judgeCount=5
```

---

## Batch Evaluation with Correctness

### Test Case 24: Batch Correctness Evaluation

**Setup:**
Create a test file `test_correctness_batch.jsonl`:

```jsonl
{"event_id": "batch-001", "event_type": "agent_response", "agent": {"name": "test"}, "interaction": {"user_query": "Capital of France?", "answer": "Paris", "expected_output": "Paris"}}
{"event_id": "batch-002", "event_type": "agent_response", "agent": {"name": "test"}, "interaction": {"user_query": "2+2?", "answer": "4", "expected_output": "Four"}}
{"event_id": "batch-003", "event_type": "agent_response", "agent": {"name": "test"}, "interaction": {"user_query": "Capital of Italy?", "answer": "Milan", "expected_output": "Rome"}}
{"event_id": "batch-004", "event_type": "agent_response", "agent": {"name": "test"}, "interaction": {"user_query": "What is Go?", "answer": "A programming language", "expected_output": "Go is a programming language created by Google"}}
```

**Run batch evaluation:**
```bash
go run cmd/batch/main.go \
  -input test_correctness_batch.jsonl \
  -output correctness_results.jsonl \
  -workers 2
```

**Expected output file `correctness_results.jsonl`:**
```jsonl
{"id":"batch-001","stages":[{"name":"correctness-judge","score":1.0,"reason":"Exact match"}],"confidence":0.95,"verdict":"pass"}
{"id":"batch-002","stages":[{"name":"correctness-judge","score":0.9,"reason":"Semantically equivalent"}],"confidence":0.88,"verdict":"pass"}
{"id":"batch-003","stages":[{"name":"correctness-judge","score":0.1,"reason":"Wrong answer - Milan is not Rome"}],"confidence":0.35,"verdict":"fail"}
{"id":"batch-004","stages":[{"name":"correctness-judge","score":0.7,"reason":"Partially correct - missing details"}],"confidence":0.68,"verdict":"review"}
```

**Console summary:**
```
INFO Starting batch evaluation input=test_correctness_batch.jsonl output=correctness_results.jsonl workers=2
INFO Worker pool finished
INFO Batch evaluation complete duration=8.234s total_records=4 success=4 errors=0
```

**Verify results:**
```bash
# Extract only correctness scores
jq -r '.stages[] | select(.name == "correctness-judge") | "\(.score) - \(.reason)"' correctness_results.jsonl

# Output:
# 1.0 - Exact match
# 0.9 - Semantically equivalent
# 0.1 - Wrong answer - Milan is not Rome
# 0.7 - Partially correct - missing details
```

**Expected:**
- All 4 records processed successfully
- Correctness judge runs for all records (all have `expected_output`)
- Scores reflect semantic similarity:
  - 1.0 for exact match (Paris = Paris)
  - 0.9 for semantic match (4 = Four)
  - 0.1 for wrong answer (Milan ≠ Rome)
  - 0.7 for partial match (missing details but correct core fact)

### Use Case: Regression Testing

**Scenario:** Test if your agent's performance changed after an update.

```bash
# 1. Create golden dataset with expected outputs
cat > golden_test_suite.jsonl << 'EOF'
{"event_id":"reg-001","event_type":"agent_response","agent":{"name":"my-agent","version":"1.0"},"interaction":{"user_query":"What is 5*6?","answer":"30","expected_output":"30"}}
{"event_id":"reg-002","event_type":"agent_response","agent":{"name":"my-agent","version":"1.0"},"interaction":{"user_query":"Capital of Spain?","answer":"Madrid","expected_output":"Madrid"}}
{"event_id":"reg-003","event_type":"agent_response","agent":{"name":"my-agent","version":"1.0"},"interaction":{"user_query":"Boiling point of water?","answer":"100°C","expected_output":"100 degrees Celsius"}}
EOF

# 2. Run evaluation
go run cmd/batch/main.go -input golden_test_suite.jsonl -output regression_v1.jsonl

# 3. Calculate average correctness score
jq -s 'map(.stages[] | select(.name == "correctness-judge") | .score) | add/length' regression_v1.jsonl
# Output: 0.95  (average correctness score)

# 4. After agent update, run again
go run cmd/batch/main.go -input golden_test_suite.jsonl -output regression_v2.jsonl
jq -s 'map(.stages[] | select(.name == "correctness-judge") | .score) | add/length' regression_v2.jsonl
# Output: 0.93  (slight regression detected)

# 5. Compare line-by-line
paste <(jq -r '.stages[] | select(.name == "correctness-judge") | .score' regression_v1.jsonl) \
      <(jq -r '.stages[] | select(.name == "correctness-judge") | .score' regression_v2.jsonl) | \
      awk '{printf "%.2f -> %.2f (%.2f change)\n", $1, $2, $2-$1}'
# Output:
# 1.00 -> 1.00 (0.00 change)
# 1.00 -> 0.95 (−0.05 change)  ← identified regression
# 0.90 -> 0.85 (−0.05 change)  ← identified regression
```

---

## Debugging Tips

### Check Server Logs

The server provides detailed logging for troubleshooting:

```bash
# Start with verbose logging
LOG_LEVEL=debug go run cmd/api/main.go
```

**Key log messages:**
```
INFO judge created successfully judge=relevance model_family=anthropic model_id=...
INFO judge completed duration=2388.927833 judge=relevance score=0.2
ERR failed to deserialize LLM response error="..." content="..."
DBG all judges completed judgeCount=5
```

### Common Issues

1. **"no clients registered for family: anthropic"**
   - **Cause:** Judge references a model not in `judges.yaml`
   - **Fix:** Add `modelFamily` and `modelID` to judge configuration

2. **"failed to deserialize LLM response"**
   - **Cause:** Model returning malformed JSON (duplicated text, extra characters)
   - **Fix:**
     - Reduce `max_tokens` (200 works well)
     - Simplify prompts
     - Enable `retry: true`
     - Use clearer output instructions: "Output ONLY valid JSON"

3. **High latency (>10s)**
   - **Cause:** Sequential judge execution or model issues
   - **Fix:** Judges run in parallel by default; check model performance

4. **Invalid threshold errors**
   - **Cause:** Threshold outside 0.0-1.0 range
   - **Fix:** Use valid threshold: `?threshold=0.7`

### Response Time Expectations

| Scenario | Expected Duration |
|----------|------------------|
| Early exit (prechecks only) | < 500ms |
| Single judge | 800ms - 2s |
| Full pipeline (5 judges) | 2.5s - 4s |
| Large context (10KB+) | 3s - 6s |

---

## Summary

**Total Test Cases:** 24

**Categories:**
- Health Check: 1 test
- Full Pipeline: 4 tests
- Single Judge: 3 tests
- Error Handling: 3 tests
- Performance: 2 tests
- Edge Cases: 2 tests
- Multi-Provider: 1 test
- Correctness Judge (API): 7 tests
- Batch Correctness: 1 test

**Expected Pass Rate:** 100% (all tests should pass with a properly configured environment)

**Note:** Correctness judge tests (17-24) require enabling the correctness judge in `configs/judges.yaml` by setting `enabled: true`.

**Configuration Requirements:**
- Valid AWS credentials (if using Bedrock)
- Valid OpenAI API key (if using GPT models)
- Properly configured `judges.yaml` with `modelFamily` and `modelID` for each judge
- All model IDs must match available models in your AWS region / OpenAI account
