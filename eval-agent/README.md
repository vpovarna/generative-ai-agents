# Eval Agent

A real-time evaluation service that scores AI agent responses using a two-stage pipeline: fast heuristic checks followed by LLM-as-Judge scoring via AWS Bedrock (Claude).

Supports two modes:
- **HTTP API** — send evaluation requests directly via REST
- **Redis Stream Consumer** — consume events from a Redis Stream for async evaluation

---

## Purpose

Given an agent's response to a user query, the eval-agent computes a structured **confidence score** and a **verdict** (`pass`, `review`, `fail`) by running:

1. **Stage 1 — PreChecks**: cheap, parallel heuristic checks (no LLM)
2. **Stage 2 — LLM Judges**: three parallel LLM calls scoring different quality dimensions
3. **Aggregation**: weighted combination of both stages into a final `EvaluationResult`

---

## Evaluation Pipeline

### Stage 1 — PreChecks (fast, no LLM)

| Checker | What it checks | Score |
|---|---|---|
| `LengthChecker` | Answer length ratio relative to query | 0.0 (too short), 0.5 (too long), 1.0 (ok) |
| `OverlapChecker` | Keyword overlap between query and answer | 0.0–1.0 based on shared unique tokens |
| `FormatChecker` | Non-empty, minimum word count, no repeated punctuation | 0.0, 0.5, or 1.0 |

If the average Stage 1 score is below the `EARLY_EXIT_THRESHOLD` (default `0.2`), the pipeline short-circuits and returns `VerdictFail` without calling the LLM — saving cost and latency.

### Stage 2 — LLM Judges (parallel, via AWS Bedrock Claude)

| Judge | What it evaluates |
|---|---|
| `RelevanceJudge` | Does the answer address the query? |
| `FaithfulnessJudge` | Is the answer grounded in the provided context (no hallucinations)? |
| `CoherenceJudge` | Is the answer internally logically consistent? |

Each judge returns a `score` (0.0–1.0) and a `reason` string parsed from a structured JSON LLM response.

### Aggregation

```
confidence = (avg_stage1 × 0.3) + (avg_stage2 × 0.7)
```

| Confidence | Verdict |
|---|---|
| > 0.8 | `pass` |
| > 0.5 | `review` |
| ≤ 0.5 | `fail` |

Weights are configurable at startup.

---

## Prerequisites

- Go 1.21+
- AWS credentials with Bedrock access
- Claude model enabled in your AWS region
- Redis (for stream consumer mode)

---

## Configuration

Create a `.env` file in `eval-agent/`:

```env
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
CLAUDE_MODEL_ID=us.anthropic.claude-3-5-haiku-20241022-v1:0
EVAL_AGENT_API_PORT=18081
EARLY_EXIT_THRESHOLD=0.2

# Redis Stream (consumer mode only)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

---

## Running

### HTTP API mode

```bash
cd eval-agent
go run cmd/api/main.go
```

Server starts on `http://localhost:18081`.

### Redis Stream consumer mode

```bash
cd eval-agent
go run cmd/main.go
```

The consumer connects to Redis, joins the `eval-group` consumer group on the `eval-events` stream, and processes messages as they arrive. Stop with `Ctrl+C` for graceful shutdown.

---

## Sending Messages to the Stream

Use `redis-cli` to publish a test message:

```bash
redis-cli XADD eval-events '*' payload '{
  "event_id": "evt-001",
  "event_type": "agent_response",
  "agent": {"name": "my-agent", "type": "rag", "version": "1.0.0"},
  "interaction": {
    "user_query": "What is the capital of France?",
    "context": "France is a country in Western Europe. Its capital city is Paris.",
    "answer": "The capital of France is Paris."
  }
}'
```

---

## API Reference

### Health Check

```bash
curl http://localhost:18081/api/v1/health
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

### Evaluate Agent Response

**POST** `/api/v1/evaluate`

```bash
curl -X POST http://localhost:18081/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-001",
    "event_type": "agent_response",
    "agent": {
      "name": "my-agent",
      "type": "rag",
      "version": "1.0.0"
    },
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe. Its capital city is Paris, which is also the largest city in the country.",
      "answer": "The capital of France is Paris."
    }
  }'
```

**Response:**
```json
{
  "id": "evt-001",
  "stages": [
    { "name": "length-checker", "score": 1.0, "reason": "Answer length is acceptable", "duration_ns": 12500 },
    { "name": "overlap-checker", "score": 0.8, "reason": "There is a good overlap", "duration_ns": 8200 },
    { "name": "format-checker", "score": 1.0, "reason": "Valid Answer", "duration_ns": 5100 },
    { "name": "relevance-judge", "score": 0.95, "reason": "The answer directly addresses the query.", "duration_ns": 820000000 },
    { "name": "faithfulness-judge", "score": 1.0, "reason": "The answer is fully supported by the context.", "duration_ns": 910000000 },
    { "name": "coherence-judge", "score": 0.95, "reason": "The answer is clear and logically consistent.", "duration_ns": 780000000 }
  ],
  "confidence": 0.965,
  "verdict": "pass"
}
```

---

### Test — Early Exit (bad answer)

```bash
curl -X POST http://localhost:18081/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-002",
    "event_type": "agent_response",
    "agent": { "name": "my-agent", "type": "rag", "version": "1.0.0" },
    "interaction": {
      "user_query": "Explain the theory of relativity in detail",
      "context": "Einstein developed the theory of relativity.",
      "answer": "ok"
    }
  }'
```

Expected: `"verdict": "fail"` with no LLM judge results (early exit triggered).
