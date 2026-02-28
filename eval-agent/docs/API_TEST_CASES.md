# API Test Cases

Comprehensive test scenarios for the eval-agent HTTP API.

## Setup

Start the API server:
```bash
cd eval-agent
go run cmd/api/main.go
```

Server runs on `http://localhost:18082`

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

### Test Case 4: Review Verdict - Moderate Quality

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
- Status Code: 200
- `confidence` between 0.5 and 0.8
- `verdict` = "review"
- 8 stages (full pipeline)
- Lower scores on relevance and faithfulness judges

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
- Status Code: 200
- Low score on `faithfulness-judge` (hallucinated China)
- Low score on `coherence-judge` (contradictory statement)
- `verdict` likely "fail"

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

## Summary

**Total Test Cases:** 15

**Categories:**
- Health Check: 1 test
- Full Pipeline: 4 tests
- Single Judge: 3 tests
- Error Handling: 3 tests
- Performance: 2 tests
- Edge Cases: 2 tests

**Expected Pass Rate:** 100% (all tests should pass with a properly configured environment)
