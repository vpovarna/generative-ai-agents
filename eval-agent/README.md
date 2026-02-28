# Eval Agent

An evaluation service that scores AI agent responses using a two-stage pipeline: fast heuristic checks followed by LLM-as-Judge scoring via AWS Bedrock (Claude).

## Purpose

Evaluates AI agent responses and returns a **confidence score** (0.0–1.0) and **verdict** (`pass`, `review`, `fail`) by analyzing:
- **Answer relevance** to the user query
- **Faithfulness** to provided context (no hallucinations)
- **Coherence** and logical consistency
- **Completeness** of the response
- **Instruction following** (format, count, style)

---

## Features

- **HTTP API** - REST endpoints for real-time evaluation
- **Batch Processing** - Evaluate datasets offline with concurrent workers
- **MCP Integration** - Use as a tool in Claude Code, Desktop, or Cursor
- **Configurable Judges** - YAML-driven prompts and model parameters
- **Early Exit** - Cost optimization with fast precheck filtering
- **Parallel Execution** - LLM judges run concurrently for speed

---

## How It Works

### Two-Stage Pipeline

```
User Query + Context + Answer
           ↓
    [Stage 1: PreChecks]     ← Fast heuristics (no LLM)
           ↓
    Early exit if score < 0.2
           ↓
    [Stage 2: LLM Judges]    ← 5 parallel Claude calls
           ↓
    [Aggregation]            ← Weighted average
           ↓
    Confidence + Verdict
```

### Stage 1: PreChecks (Fast, No LLM)

| Checker | Checks | Output |
|---------|--------|--------|
| **LengthChecker** | Answer/query length ratio | 0.0 (too short), 0.5 (too long), 1.0 (ok) |
| **OverlapChecker** | Keyword overlap | 0.0–1.0 based on shared tokens |
| **FormatChecker** | Non-empty, word count, punctuation | 0.0, 0.5, or 1.0 |

**Early exit:** If average Stage 1 score < 0.2, returns `fail` verdict without calling LLM (saves cost/latency).

### Stage 2: LLM Judges (Parallel, AWS Bedrock Claude)

**Configurable via YAML** - All judges are loaded from `configs/judges.yaml` allowing prompt customization without code changes.

| Judge | Evaluates | Scoring Rubric |
|-------|-----------|----------------|
| **relevance** | Does answer address the query? | 1.0 (highly relevant) → 0.0 (unrelated) |
| **faithfulness** | Grounded in context? (no hallucinations) | 1.0 (all grounded) → 0.0 (mostly hallucinated) |
| **coherence** | Internally consistent logic? | 1.0 (fully coherent) → 0.0 (contradictory) |
| **completeness** | Fully addresses all parts of query? | 1.0 (all addressed), 0.5 (some missing), 0.0 (major parts ignored) |
| **instruction** | Follows explicit instructions? (format, count, style) | 1.0 (all followed), 0.7-0.9 (most), 0.4-0.6 (some), 0.0-0.3 (mostly ignored) |

Each judge returns `score` (0.0–1.0) + `reason` string.

**Performance:**
- Judges run in **parallel** for speed
- 15-second timeout per judge
- Automatic retry with exponential backoff

### Aggregation

```
confidence = (avg_stage1 × 0.3) + (avg_stage2 × 0.7)
```

| Confidence | Verdict |
|------------|---------|
| > 0.8 | `pass` |
| > 0.5 | `review` |
| ≤ 0.5 | `fail` |

---

## Quick Start

### Prerequisites

- Go 1.21+
- AWS credentials with Bedrock access (Claude enabled)

### Configuration

Create `.env` in `eval-agent/`:

```env
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
CLAUDE_MODEL_ID=us.anthropic.claude-3-5-haiku-20241022-v1:0
EVAL_AGENT_API_PORT=18082
EARLY_EXIT_THRESHOLD=0.2
```

Judges are configured in `configs/judges.yaml` - see [Judge Configuration](#judge-configuration) section.

---

## Usage Modes

### 1. HTTP API (Real-time)

REST endpoints for real-time evaluation of agent responses. Supports full pipeline evaluation with all judges or single-judge evaluation for faster results.

**Key capabilities:**
- Full pipeline with prechecks + all LLM judges
- Single judge evaluation with custom thresholds
- Health check endpoint for monitoring

📚 **Documentation:** [docs/API_TEST_CASES.md](docs/API_TEST_CASES.md)

### 2. Batch Processing (Offline)

CLI tool for evaluating datasets offline with concurrent workers. Processes JSONL files and outputs results in multiple formats.

**Key capabilities:**
- Concurrent evaluation with configurable worker pool
- JSONL and summary output formats
- Graceful shutdown and dry-run validation
- Progress tracking with timing

📦 **Documentation:** [docs/BATCH_EVALUATION.md](docs/BATCH_EVALUATION.md)

### 3. MCP Integration (Claude Code/Desktop/Cursor)

Expose eval-agent as a tool in Claude Code, Claude Desktop, or Cursor. Enables Claude to evaluate agent responses directly during conversations.

**Key capabilities:**
- Two tools: `evaluate_response` (full pipeline) and `evaluate_single_judge`
- Works with Claude Code, Claude Desktop, and Cursor
- Docker and binary deployment options

![eval-agent MCP tool in Claude Code](docs/image.png)

🔌 **Documentation:** [docs/MCP_TEST_CASES.md](docs/MCP_TEST_CASES.md)

---

## API Reference

### Full Pipeline Evaluation

**POST** `/api/v1/evaluate`

Runs both stages (prechecks + all LLM judges) and returns aggregated result.

### Single Judge Evaluation

**POST** `/api/v1/evaluate/judge/{judge_name}?threshold=0.7`

Evaluates with only one judge. Available judges: `relevance`, `faithfulness`, `coherence`, `completeness`, `instruction`

**Query params:**
- `threshold` (optional): Pass/fail threshold (0.0-1.0, default: 0.7)

**Example:**
```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/relevance?threshold=0.9" \
  -H "Content-Type: application/json" \
  -d '{...}'
```

---

## Judge Configuration

Judges are defined in `configs/judges.yaml`:

```yaml
judges:
  default_model:
    max_tokens: 256
    temperature: 0.0
    retry: true

  evaluators:
    - name: relevance
      enabled: true
      requires_context: false
      model:
        max_tokens: 256
        temperature: 0.0
        retry: false
      prompt: |
        You are an evaluation judge.
        Score how relevant the answer is to the query...

        Query: {{.Query}}
        Answer: {{.Answer}}

        {"score": <float>, "reason": "<string>"}
```

**Benefits:**
- ✅ Edit prompts without code changes
- ✅ Enable/disable judges per deployment
- ✅ Override model settings per judge
- ✅ A/B test different configurations

---

## Testing

Comprehensive test cases and setup instructions for each usage mode:

| Test Suite | Description | Documentation |
|------------|-------------|---------------|
| **API Tests** | HTTP endpoints, error handling, edge cases | [docs/API_TEST_CASES.md](docs/API_TEST_CASES.md) |
| **Batch Tests** | JSONL processing, workers, formats | [docs/BATCH_EVALUATION.md](docs/BATCH_EVALUATION.md) |
| **MCP Tests** | Tool integration, Claude Code/Desktop/Cursor | [docs/MCP_TEST_CASES.md](docs/MCP_TEST_CASES.md) |
| **Legacy Tests** | Original test dataset and examples | [docs/TESTING.md](docs/TESTING.md) |

---

## License

MIT
