# Eval Agent

A real-time evaluation service that scores AI agent responses using a two-stage pipeline: fast heuristic checks followed by LLM-as-Judge scoring via AWS Bedrock (Claude).

Supports three modes:
- **HTTP API** — send evaluation requests directly via REST
- **Redis Stream Consumer** — consume events from a Redis Stream for async evaluation
- **MCP (Model Context Protocol)** — expose evaluation as a tool for Cursor, Claude Desktop, and other MCP clients

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
| `CompletenessJudge` | Does the answer fully address all distinct questions/sub-requests in the query? |
| `InstructionJudge` | Does the answer follow explicit instructions (format, count, style, etc.)? |

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

### MCP mode (stdio)

```bash
cd eval-agent
go run cmd/mcp/main.go
```

The server runs over stdio and exposes the `evaluate_response` tool. Configure it in Cursor (Settings → MCP) or Claude Desktop (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "eval-agent": {
      "command": "go",
      "args": ["run", "cmd/mcp/main.go"],
      "cwd": "/path/to/eval-agent"
    }
  }
}
```

**Tool input** (matches HTTP API field names):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `event_id` | string | yes | Unique event identifier |
| `user_query` | string | yes | User's original query |
| `answer` | string | yes | Agent response to evaluate |
| `context` | string | no | Optional context or retrieved documents |

Stop with `Ctrl+C` for graceful shutdown.

#### Test manually (JSON-RPC)

Build the binary first:

```bash
cd eval-agent
go build -o bin/eval-mcp cmd/mcp/main.go
```

MCP requires an **initialize handshake** before `tools/list` or `tools/call`. Send these messages in order (one JSON-RPC message per line). Use `sleep 2` to keep the pipe open so the server can respond before EOF.

Test `tools/list`:

```bash
(printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  ; sleep 2) | ./bin/eval-mcp 2>/dev/null
```

Expected output: JSON with `result.tools` containing `evaluate_response`.

Test `evaluate_response`:

```bash
(printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"evaluate_response","arguments":{"event_id":"test-1","user_query":"What is AI?","answer":"AI is artificial intelligence","context":""}}}' \
  ; sleep 5) | ./bin/eval-mcp 2>/dev/null
```

Expected output: JSON with confidence score, verdict, and all stage results.

#### Add to Claude Code

```bash
claude mcp add --transport stdio --scope project eval-agent \
  -- /path/to/eval-agent/bin/eval-mcp
```

Replace `/path/to/eval-agent` with your actual eval-agent directory.

Verify configuration:

```bash
claude mcp list
```

Should show: `eval-agent (stdio) - Ready`

#### Test in Claude Code session

Start a new Claude Code session and ask:

> Use the evaluate_response tool to evaluate this: Query='What is RAG?', Answer='RAG is Retrieval Augmented Generation, a technique that combines information retrieval with text generation.', Context='RAG systems retrieve relevant documents and use them to generate accurate responses.'

Claude will call the tool and return something like: `{"confidence": 0.92, "verdict": "pass", "stages": [...]}`

#### View MCP status

Within a Claude Code session, run:

```
/mcp
```

Shows all connected MCP servers and their tools.

---

## Sending Messages to the Stream

### CLI Producer (recommended)

```bash
go run cmd/producer/main.go -d '{
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

Flags: `-d` (inline JSON), `--redis-addr`, `--stream` (default: eval-events)

### Alternative: redis-cli

```bash
redis-cli XADD eval-events '*' payload '{"event_id":"evt-001","event_type":"agent_response","agent":{"name":"my-agent","type":"rag","version":"1.0.0"},"interaction":{"user_query":"What is the capital of France?","context":"France is a country in Western Europe. Its capital city is Paris.","answer":"The capital of France is Paris."}}'
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

### Test Completeness Judge
```bash
curl -X POST http://localhost:18081/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-completeness",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Explain both encryption and decryption, and provide examples of each",
      "context": "Encryption converts plaintext to ciphertext. Decryption reverses this.",
      "answer": "Encryption converts plaintext to ciphertext using a key."
    }
  }'
```

Expected: `"verdict": "fail"` Low completeness score (only addressed encryption, missed decryption and examples)


### Test Instruction judge
```bash
  curl -X POST http://localhost:18081/api/v1/evaluate \
    -H "Content-Type: application/json" \
    -d '{
      "event_id": "test-off-by-one",
      "event_type": "agent_response",
      "agent": {"name": "test", "type": "rag", "version": "1.0"},
      "interaction": {
        "user_query": "Give me 3 examples of encryption algorithms",
        "context": "Common algorithms: AES, RSA, ChaCha20, Blowfish",
        "answer": "Examples of encryption algorithms are AES, RSA, ChaCha20, and Blowfish."
      }
    }' | jq '.stages[] | select(.name=="instruction-judge")'
```

Expected: `"verdict": "pass"`. asked for 3, gave 4 - minor overshoot

---

## Single Judge Evaluation

Evaluate using only one specific judge, bypassing the full pipeline. Useful for targeted testing or when you only need one quality dimension.

**POST** `/api/v1/evaluate/judge/{judge_name}?threshold=0.7`

**Available judges:** `relevance`, `faithfulness`, `coherence`, `completeness`, `instruction`

**Query parameters:**
- `threshold` (optional): Pass/fail threshold (0.0-1.0, default: 0.7)

### Test Relevance Judge (default threshold)

```bash
curl -X POST http://localhost:18081/api/v1/evaluate/judge/relevance \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-relevance",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe. Paris is its capital.",
      "answer": "The capital of France is Paris."
    }
  }'
```

**Expected response:**
```json
{
  "id": "test-relevance",
  "stages": [
    {
      "name": "relevance-judge",
      "score": 0.95,
      "reason": "The answer directly addresses the query about France's capital.",
      "duration_ns": 850000000
    }
  ],
  "confidence": 0.95,
  "verdict": "pass"
}
```

### Test Faithfulness Judge (custom threshold)

```bash
curl -X POST "http://localhost:18081/api/v1/evaluate/judge/faithfulness?threshold=0.9" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-faithfulness",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What encryption does the product support?",
      "context": "Our product supports AES-256 encryption for data at rest.",
      "answer": "The product supports AES-256 encryption and also offers quantum-resistant algorithms."
    }
  }'
```

**Expected:** Low score (hallucination detected - quantum-resistant not in context), `"verdict": "fail"` with threshold 0.9

### Test Coherence Judge

```bash
curl -X POST http://localhost:18081/api/v1/evaluate/judge/coherence \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-coherence",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "How does encryption work?",
      "context": "Encryption uses mathematical algorithms to scramble data.",
      "answer": "Encryption scrambles data. But also, pizza is delicious. The algorithm ensures security."
    }
  }'
```

**Expected:** Low score (incoherent - random pizza statement), `"verdict": "fail"`

### Test Completeness Judge

```bash
curl -X POST http://localhost:18081/api/v1/evaluate/judge/completeness \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-completeness",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Explain both encryption and decryption, and provide examples",
      "context": "Encryption converts plaintext to ciphertext. Decryption reverses this. Example: AES-256.",
      "answer": "Encryption converts plaintext to ciphertext using algorithms like AES-256."
    }
  }'
```

**Expected:** Low score (incomplete - missed decryption and full examples), `"verdict": "fail"`

### Test Instruction Judge

```bash
curl -X POST http://localhost:18081/api/v1/evaluate/judge/instruction \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-instruction",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "List exactly 3 encryption algorithms",
      "context": "Common algorithms: AES, RSA, ChaCha20, Blowfish, Twofish",
      "answer": "1. AES\n2. RSA"
    }
  }'
```

**Expected:** Low score (instruction violation - asked for 3, provided only 2), `"verdict": "fail"`

### Test with Strict Threshold

```bash
curl -X POST "http://localhost:18081/api/v1/evaluate/judge/relevance?threshold=0.95" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-strict",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is AI?",
      "context": "AI stands for Artificial Intelligence.",
      "answer": "AI refers to artificial intelligence, which is technology that mimics human cognition."
    }
  }'
```

**Expected:** High relevance score (~0.9-0.95), verdict depends on whether score exceeds 0.95 threshold
