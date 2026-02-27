# Eval Agent Implementation Roadmap

## Overview

This document outlines the phased implementation plan to evolve eval-agent from a **real-time evaluation API** into a comprehensive **evaluation platform** that supports both production monitoring and research-grade validation workflows.

**Current State:**
- Real-time HTTP API for single-request evaluation
- Two-stage pipeline (prechecks + 5 LLM judges)
- Early-exit optimization for cost savings
- MCP integration for Claude Code
- Redis Stream support for async processing

**Target State:**
- Batch dataset evaluation CLI
- YAML-driven configurable judges (no code changes)
- Human annotation validation workflow
- Correlation analysis (Kendall's tau)
- Iterative prompt improvement loop
- Reference-based vs reference-free modes

---

## Phase 1: Configuration System (YAML-Driven Judges)

**Goal:** Make judges configurable without code changes using YAML templates.

### 1.1 Implementation Tasks

**Task 1.1.1: Design YAML Schema**
```yaml
# configs/default.yaml
llm_judge:
  prompt: |
    You are an expert evaluator assessing the correctness of responses to user queries.

    ## Task
    Evaluate whether the response correctly answers the user's query.

  annotation_labels:
    - score: 0
      label: "incorrect"
      description: "The response does not answer the query correctly or contains significant factual errors"
    - score: 1
      label: "partial"
      description: "The response addresses the query but is missing important details or contains minor inaccuracies"
    - score: 2
      label: "correct"
      description: "The response fully and accurately answers the query with all necessary information"

# Support for different evaluation types
evaluation_type: "reference"  # or "no_reference"

# Optional: Judge-specific configurations
judges:
  relevance:
    enabled: true
    weight: 0.25
    prompt_override: |
      Evaluate if the answer addresses the user's query...

  faithfulness:
    enabled: true
    weight: 0.25
    require_context: true

  coherence:
    enabled: true
    weight: 0.20

  completeness:
    enabled: true
    weight: 0.15

  instruction:
    enabled: true
    weight: 0.15
```

**Task 1.1.2: Create Config Loader** (`internal/config/`)
- [ ] `loader.go` - Load and validate YAML configs
- [ ] `schema.go` - Config struct definitions
- [ ] `validator.go` - Validate annotation labels are contiguous from 0
- [ ] `loader_test.go` - Unit tests for config loading

```go
type Config struct {
    LLMJudge        LLMJudgeConfig     `yaml:"llm_judge"`
    EvaluationType  string             `yaml:"evaluation_type"`
    Judges          map[string]JudgeConfig `yaml:"judges"`
}

type LLMJudgeConfig struct {
    Prompt           string            `yaml:"prompt"`
    AnnotationLabels []AnnotationLabel `yaml:"annotation_labels"`
}

type AnnotationLabel struct {
    Score       int    `yaml:"score"`
    Label       string `yaml:"label"`
    Description string `yaml:"description"`
}

type JudgeConfig struct {
    Enabled        bool    `yaml:"enabled"`
    Weight         float64 `yaml:"weight"`
    PromptOverride string  `yaml:"prompt_override"`
    RequireContext bool    `yaml:"require_context"`
}
```

**Task 1.1.3: Refactor Judges to Use Config**
- [ ] Update `internal/judge/judge.go` base interface to accept config
- [ ] Refactor all 5 judges to use dynamic prompts from config
- [ ] Add prompt template interpolation (inject query, answer, context)
- [ ] Update judge factory to pass config to constructors

**Task 1.1.4: CLI Flag for Config Path**
- [ ] Add `--config` flag to `cmd/api/main.go`
- [ ] Add `--config` flag to `cmd/mcp/main.go`
- [ ] Default to `configs/default.yaml` if not specified
- [ ] Add config validation on startup

**Task 1.1.5: Sample Configs**
Create example configs in `configs/`:
- [ ] `default.yaml` - General correctness evaluation
- [ ] `helpfulness.yaml` - Customer support quality
- [ ] `technical-accuracy.yaml` - Coding/technical responses
- [ ] `creative.yaml` - Creative writing (no reference mode)

### 1.2 Testing Requirements

- [ ] Unit tests for YAML parsing and validation
- [ ] Integration tests with each sample config
- [ ] Verify judges produce expected output with custom prompts
- [ ] Test invalid configs (non-contiguous scores, missing fields)

### 1.3 Expected Outcomes

- ✅ Users can create custom evaluation criteria without modifying code
- ✅ Multiple evaluation profiles for different use cases
- ✅ Easier iteration on judge prompts

### 1.4 Dependencies

None (can implement immediately)

### 1.5 Estimated Effort

**5-7 days** (1 engineer)

---

## Phase 2: Batch Processing CLI

**Goal:** Add CLI for batch evaluation of datasets (JSON files).

### 2.1 Implementation Tasks

**Task 2.1.1: Dataset Schema** (`internal/batch/schema.go`)

```go
type DatasetEntry struct {
    UUID           string `json:"uuid"`
    Skill          string `json:"skill"`
    OperatorName   string `json:"operator_name"`
    Input          string `json:"input"`
    Predictions    string `json:"predictions"`
    ExpectedOutput string `json:"expected_output,omitempty"` // for reference-based
}

type Dataset []DatasetEntry

type EvaluationOutput struct {
    UUID       string               `json:"uuid"`
    Input      string               `json:"input"`
    Prediction string               `json:"prediction"`
    Confidence float64              `json:"confidence"`
    Verdict    string               `json:"verdict"`
    Stages     []models.StageResult `json:"stages"`
    Duration   time.Duration        `json:"duration"`
}

type BatchResults []EvaluationOutput
```

**Task 2.1.2: Batch Processor** (`internal/batch/processor.go`)
- [ ] Load JSON dataset from file
- [ ] Validate required fields (`uuid`, `input`, `predictions`)
- [ ] Process entries sequentially or with worker pool
- [ ] Progress bar / logging for long-running batches
- [ ] Write results to `{input_file}.llm-judge-results.json`

```go
type Processor struct {
    executor   *executor.Executor
    config     *config.Config
    logger     *zerolog.Logger
    workerPool int // parallel workers
}

func (p *Processor) ProcessDataset(
    ctx context.Context,
    datasetPath string,
    outputPath string,
) error {
    // 1. Load dataset
    // 2. For each entry, call executor.Execute()
    // 3. Write results incrementally (streaming JSON)
    // 4. Return summary stats
}
```

**Task 2.1.3: CLI Tool** (`cmd/batch/main.go`)
```bash
# Basic usage
go run cmd/batch/main.go \
  --data-file=datasets/benchmark.v1.json \
  --config=configs/default.yaml \
  --output=results/benchmark.v1.llm-judge-results.json

# Optional flags
go run cmd/batch/main.go \
  --data-file=datasets/benchmark.v1.json \
  --config=configs/default.yaml \
  --evaluation-type=no_reference \
  --workers=5 \
  --early-exit-threshold=0.2
```

**Flags:**
- `--data-file` (required): Input JSON dataset
- `--config` (optional): YAML config path (default: `configs/default.yaml`)
- `--output` (optional): Output path (default: `{data_file}.llm-judge-results.json`)
- `--evaluation-type` (optional): `reference` or `no_reference` (default from config)
- `--workers` (optional): Parallel workers (default: 5)
- `--early-exit-threshold` (optional): Override config value

**Task 2.1.4: Result Aggregation**
- [ ] Generate summary statistics after batch completes
- [ ] Write to `{input_file}.summary.json`:
  ```json
  {
    "total_entries": 500,
    "pass": 350,
    "review": 100,
    "fail": 50,
    "avg_confidence": 0.78,
    "avg_duration_ms": 1200,
    "early_exits": 45
  }
  ```

**Task 2.1.5: Makefile Integration**
```makefile
# Makefile in eval-agent/

.PHONY: run-batch
run-batch:
	@go run cmd/batch/main.go \
		--data-file=$(DATA_FILE) \
		--config=$(CONFIG) \
		--evaluation-type=$(EVAL_TYPE)

# Example usage:
# make run-batch DATA_FILE=datasets/test.json CONFIG=configs/default.yaml EVAL_TYPE=reference
```

### 2.2 Testing Requirements

- [ ] Unit tests for dataset loading and validation
- [ ] Integration test with sample dataset (10-20 entries)
- [ ] Test invalid JSON, missing fields, malformed entries
- [ ] Test worker pool parallelism
- [ ] Test incremental output writing (crash recovery)

### 2.3 Expected Outcomes

- ✅ Can evaluate 100s-1000s of entries in batch
- ✅ Structured JSON output for downstream analysis
- ✅ Progress tracking for long-running jobs
- ✅ Parallel processing for speed

### 2.4 Dependencies

- ✅ Phase 1 (YAML config system)

### 2.5 Estimated Effort

**7-10 days** (1 engineer)

---

## Phase 3: Human Annotation Workflow

**Goal:** Support human-annotated datasets for judge validation.

### 3.1 Implementation Tasks

**Task 3.1.1: Human Annotation Schema** (`internal/batch/human_annotations.go`)

```go
type HumanAnnotatedEntry struct {
    UUID             string `json:"uuid"`
    Input            string `json:"input"`
    Skill            string `json:"skill"`
    OperatorName     string `json:"operator_name"`
    Predictions      string `json:"predictions"`
    HumanAnnotation  string `json:"human_annotation"` // matches config labels
}

type HumanAnnotations []HumanAnnotatedEntry
```

**Task 3.1.2: Annotation Validator** (`internal/batch/annotation_validator.go`)
- [ ] Load human annotations JSON
- [ ] Validate `human_annotation` values match config `annotation_labels.label`
- [ ] Check for missing annotations
- [ ] Return validation errors

**Task 3.1.3: Sampling Logic** (`internal/batch/sampler.go`)
- [ ] Sample 25% of dataset entries
- [ ] Stratified sampling (if skills/operators present)
- [ ] Write sampled entries to `{dataset}_sampled.json`
- [ ] Generate annotation template for humans:
  ```json
  [
    {
      "uuid": "uuid-123",
      "input": "What is AI?",
      "predictions": "AI is...",
      "human_annotation": ""  // to be filled
    }
  ]
  ```

**Task 3.1.4: CLI Integration**
```bash
# Sample dataset for human annotation
go run cmd/batch/main.go sample \
  --data-file=datasets/benchmark.json \
  --output=datasets/benchmark_sampled.json \
  --sample-rate=0.25

# Validate human annotations against config
go run cmd/batch/main.go validate-annotations \
  --annotations=human_annotations.json \
  --config=configs/default.yaml
```

### 3.2 Testing Requirements

- [ ] Test sampling with different rates (0.25, 0.50)
- [ ] Test stratified sampling by skill/operator
- [ ] Test annotation validation (valid/invalid labels)
- [ ] Test annotation format compatibility

### 3.3 Expected Outcomes

- ✅ Generate human annotation templates from datasets
- ✅ Validate annotation files before correlation analysis
- ✅ Support integration with annotation platforms (Scale AI, etc.)

### 3.4 Dependencies

- ✅ Phase 2 (Batch CLI)

### 3.5 Estimated Effort

**3-5 days** (1 engineer)

---

## Phase 4: Correlation Analysis

**Goal:** Compute correlation between LLM judges and human annotations.

### 4.1 Implementation Tasks

**Task 4.1.1: Correlation Metrics** (`internal/correlation/metrics.go`)

```go
type CorrelationResult struct {
    KendallTau     float64 `json:"kendall_tau"`
    PValue         float64 `json:"p_value"`
    SampleSize     int     `json:"sample_size"`
    LLMScores      []float64 `json:"llm_scores"`
    HumanScores    []float64 `json:"human_scores"`
}

// Compute Kendall's Tau correlation
func ComputeKendallTau(llmScores, humanScores []float64) CorrelationResult
```

Use library: `github.com/gonum/stat` or implement manually.

**Task 4.1.2: Semantic Similarity** (`internal/correlation/similarity.go`)

```go
// Check if sampled data distribution matches full dataset
type SimilarityResult struct {
    CosineSimilarity float64 `json:"cosine_similarity"`
    MeetsThreshold   bool    `json:"meets_threshold"` // >= 0.8
    SampleSize       int     `json:"sample_size"`
}

func ComputeSemanticSimilarity(
    fullDataset []DatasetEntry,
    sampledDataset []DatasetEntry,
    embeddingClient *embedding.Client,
) SimilarityResult
```

**Task 4.1.3: Correlation Analyzer** (`internal/correlation/analyzer.go`)
- [ ] Load LLM judge results from batch evaluation
- [ ] Load human annotations
- [ ] Match by UUID
- [ ] Map human annotation labels to numeric scores using config
- [ ] Compute Kendall's Tau
- [ ] Write results to `{dataset}.llm-judge-correlation.json`

```go
type Analyzer struct {
    config *config.Config
    logger *zerolog.Logger
}

func (a *Analyzer) Analyze(
    llmResults BatchResults,
    humanAnnotations HumanAnnotations,
) CorrelationResult
```

**Task 4.1.4: CLI Integration**
```bash
# Run correlation analysis
go run cmd/batch/main.go correlate \
  --llm-results=datasets/benchmark.llm-judge-results.json \
  --human-annotations=human_annotations.json \
  --config=configs/default.yaml \
  --threshold=0.3
```

**Output:**
```json
{
  "kendall_tau": 0.65,
  "p_value": 0.001,
  "sample_size": 125,
  "meets_threshold": true,
  "threshold": 0.3,
  "scatter_plot_data": {
    "llm_scores": [0.9, 0.8, ...],
    "human_scores": [2, 2, ...]
  }
}
```

**Task 4.1.5: Threshold-Based Decisions**
- [ ] If correlation < threshold (e.g., 0.3), print warning:
  ```
  ⚠️  Kendall's correlation (0.25) is below threshold of 0.3.

  Recommendations:
  1. Improve your evaluation prompt in configs/default.yaml
  2. Run: go run cmd/batch/main.go iterate ...
  3. Re-run correlation analysis after prompt improvements
  ```

- [ ] If correlation ≥ threshold, proceed to full dataset evaluation

**Task 4.1.6: Visualization Data Export**
- [ ] Export scatter plot data (LLM vs human scores)
- [ ] Export confusion matrix data
- [ ] Format for easy plotting (Python/matplotlib, JS/plotly)

### 4.2 Testing Requirements

- [ ] Unit tests for Kendall's Tau computation
- [ ] Test with perfect correlation (tau = 1.0)
- [ ] Test with no correlation (tau ≈ 0.0)
- [ ] Test with synthetic LLM + human scores
- [ ] Integration test with sample annotations

### 4.3 Expected Outcomes

- ✅ Quantitative measure of LLM judge quality
- ✅ Data-driven decision on judge reliability
- ✅ Visualization-ready output for analysis

### 4.4 Dependencies

- ✅ Phase 2 (Batch CLI)
- ✅ Phase 3 (Human Annotations)

### 4.5 Estimated Effort

**5-7 days** (1 engineer)

---

## Phase 5: Iterative Improvement Loop

**Goal:** Enable iterative prompt engineering to improve judge quality.

### 5.1 Implementation Tasks

**Task 5.1.1: Iterate Command** (`cmd/batch/main.go`)

```bash
# Iterate on prompt quality with human-annotated subset
go run cmd/batch/main.go iterate \
  --human-annotations=human_annotations.json \
  --config=configs/default.yaml \
  --threshold=0.3
```

**Workflow:**
1. Load human annotations
2. Run LLM judges on annotated subset (not full dataset)
3. Compute correlation
4. Display results + recommendations
5. Exit (user edits config, re-runs)

**Task 5.1.2: Feedback Loop**
- [ ] Display current correlation
- [ ] Show examples where LLM disagrees with humans
- [ ] Suggest prompt improvements based on mismatches
- [ ] Track iteration history (optional)

```
Iteration 1: Kendall's Tau = 0.28 (below threshold 0.3)

Top Disagreements:
- UUID: abc-123
  Human: "incorrect" (score 0)
  LLM: 0.85 (predicted "correct")
  Query: "What is 2+2?"
  Answer: "5"

Suggestion: Add explicit instruction to check mathematical correctness.
```

**Task 5.1.3: Full Workflow Integration** (`run-llm-judge` command)

```bash
# Full LLM-as-Judge workflow with decision points
go run cmd/batch/main.go run-llm-judge \
  --data-file=datasets/benchmark.json \
  --config=configs/default.yaml \
  --human-annotations=human_annotations.json \
  --evaluation-type=reference
```

**Workflow Logic:**
1. If no human annotations provided → sample + generate annotation template
2. If human annotations provided → run correlation analysis
3. If correlation < 0.3 → suggest iterate command, exit
4. If correlation ≥ 0.3 → run full dataset evaluation
5. Sample 25% for semantic similarity check
6. If similarity < 0.8 → suggest additional annotations, exit
7. If similarity ≥ 0.8 → output final results

**Task 5.1.4: Makefile Commands**

```makefile
# Run full LLM-as-Judge workflow
.PHONY: run-llm-judge
run-llm-judge:
	@go run cmd/batch/main.go run-llm-judge \
		--data-file=$(DATA_FILE) \
		--config=$(CONFIG_PATH) \
		--evaluation-type=$(EVALUATOR_TYPE) \
		$(if $(HUMAN_ANNOTATIONS),--human-annotations=$(HUMAN_ANNOTATIONS))

# Iterate on prompt quality
.PHONY: iterate-llm-judge
iterate-llm-judge:
	@go run cmd/batch/main.go iterate \
		--human-annotations=$(HUMAN_ANNOTATIONS) \
		--config=$(CONFIG_PATH) \
		--evaluation-type=$(EVALUATOR_TYPE)

# Example usage:
# make run-llm-judge DATA_FILE=benchmark.json CONFIG_PATH=configs/default.yaml EVALUATOR_TYPE=reference
# make iterate-llm-judge HUMAN_ANNOTATIONS=labels.json CONFIG_PATH=configs/default.yaml EVALUATOR_TYPE=reference
```

### 5.2 Testing Requirements

- [ ] End-to-end test of full workflow
- [ ] Test decision points (correlation thresholds)
- [ ] Test with low/high correlation scenarios
- [ ] Test semantic similarity checks

### 5.3 Expected Outcomes

- ✅ Closed-loop prompt improvement workflow
- ✅ Data-driven iteration on judge quality
- ✅ Complete validation pipeline before production deployment

### 5.4 Dependencies

- ✅ Phase 4 (Correlation Analysis)

### 5.5 Estimated Effort

**5-7 days** (1 engineer)

---

## Phase 6: Production Integration (Optional)

**Goal:** Bridge batch evaluation with real-time API.

### 6.1 Implementation Tasks

**Task 6.1.1: Judge Model Registry**
- [ ] Store validated judge configs in database/Redis
- [ ] API to fetch validated configs: `GET /api/v1/judges/{name}`
- [ ] API uses pre-validated configs from batch workflow

**Task 6.1.2: A/B Testing Support**
- [ ] Route % of traffic to new judge configs
- [ ] Compare performance metrics (confidence distribution, latency)
- [ ] Gradual rollout of improved judges

**Task 6.1.3: Continuous Monitoring**
- [ ] Export real-time eval results to batch format
- [ ] Periodic correlation analysis against human feedback
- [ ] Alert if correlation drops below threshold

### 6.2 Dependencies

- ✅ All previous phases

### 6.3 Estimated Effort

**7-10 days** (1 engineer)

---

## Summary Timeline

| Phase | Description | Effort | Dependencies |
|-------|-------------|--------|--------------|
| **Phase 1** | YAML Config System | 5-7 days | None |
| **Phase 2** | Batch Processing CLI | 7-10 days | Phase 1 |
| **Phase 3** | Human Annotations | 3-5 days | Phase 2 |
| **Phase 4** | Correlation Analysis | 5-7 days | Phase 2, 3 |
| **Phase 5** | Iterative Improvement | 5-7 days | Phase 4 |
| **Phase 6** | Production Integration | 7-10 days | All |

**Total Estimated Effort:** 32-46 days (1 engineer) or **6-9 weeks**

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        EVAL AGENT                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Production Mode                    Research Mode            │
│  ┌─────────────────┐               ┌──────────────────┐    │
│  │  HTTP API       │               │  Batch CLI       │    │
│  │  (cmd/api)      │               │  (cmd/batch)     │    │
│  │  - /evaluate    │               │  - sample        │    │
│  │  - /judge/{id}  │               │  - validate      │    │
│  │  - /health      │               │  - correlate     │    │
│  └────────┬────────┘               │  - iterate       │    │
│           │                         │  - run-llm-judge │    │
│           │                         └────────┬─────────┘    │
│  ┌────────▼────────┐                        │               │
│  │  Redis Stream   │                        │               │
│  │  (cmd/main.go)  │                        │               │
│  └────────┬────────┘                        │               │
│           │                                  │               │
│           └────────────┬─────────────────────┘               │
│                        │                                     │
│                ┌───────▼──────────┐                          │
│                │  Executor        │                          │
│                │  (orchestration) │                          │
│                └───────┬──────────┘                          │
│                        │                                     │
│         ┌──────────────┼──────────────┐                     │
│         │              │               │                     │
│  ┌──────▼──────┐ ┌────▼─────┐ ┌──────▼──────┐             │
│  │  PreChecks  │ │  Judges  │ │ Aggregator  │             │
│  │  - length   │ │ (YAML)   │ │ - weighted  │             │
│  │  - overlap  │ │ - config │ │ - verdict   │             │
│  │  - format   │ │ - dynamic│ │             │             │
│  └─────────────┘ └──────────┘ └─────────────┘             │
│                        │                                     │
│                 ┌──────▼─────────┐                          │
│                 │  AWS Bedrock   │                          │
│                 │  (Claude)      │                          │
│                 └────────────────┘                          │
│                                                              │
│  Research Tools                                              │
│  ┌─────────────────────────────────────────────────┐        │
│  │  Correlation Analyzer                           │        │
│  │  - Kendall's Tau                                │        │
│  │  - Semantic Similarity                          │        │
│  │  - Human Annotation Validation                  │        │
│  └─────────────────────────────────────────────────┘        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Decision: CLI vs API Endpoint

**Question:** Should batch evaluation be a CLI or HTTP API endpoint?

**Recommendation:** **CLI** (`cmd/batch/main.go`)

**Rationale:**
1. ✅ **Long-running jobs:** Batch eval can take minutes/hours → bad for HTTP
2. ✅ **File I/O:** Reads/writes large JSON files → better suited for CLI
3. ✅ **Human workflow:** Researchers iterate locally, not via API
4. ✅ **Separation of concerns:** API for production, CLI for research
5. ✅ **Makefile orchestration:** Easier to script with Make than curl

**Alternative:** If async processing needed, use Redis Stream (already implemented) or add job queue (Celery, Temporal).

---

## Open Questions

1. **Annotation Platform Integration:**
   - Do we integrate with Scale AI, Label Studio, or other platforms?
   - Or export generic JSON and let users handle annotation externally?
   - **Recommendation:** Start with generic JSON export, add integrations later.

2. **Embedding Model for Semantic Similarity:**
   - Use AWS Bedrock embeddings (Titan, Claude)?
   - Use OpenAI embeddings?
   - Use local embeddings (sentence-transformers)?
   - **Recommendation:** Use existing Bedrock embeddings from kg-agent.

3. **Regression Testing:**
   - Should we add dataset versioning (benchmark.v1.json, benchmark.v2.json)?
   - Track metric deltas between agent versions?
   - **Recommendation:** Phase 7 (future work).

4. **Multi-Language Support:**
   - Should configs support multiple languages (i18n)?
   - **Recommendation:** Not in initial phases, add later if needed.

---

## Success Metrics

**Phase 1 Success:**
- [ ] Can evaluate with custom YAML config
- [ ] No code changes needed for new evaluation criteria

**Phase 2 Success:**
- [ ] Can batch evaluate 500+ entries
- [ ] Results written to JSON in < 30 minutes

**Phase 3 Success:**
- [ ] Can sample dataset and generate annotation template
- [ ] Can validate human annotations against config

**Phase 4 Success:**
- [ ] Can compute Kendall's Tau correlation
- [ ] Correlation threshold gates full dataset evaluation

**Phase 5 Success:**
- [ ] Iterative workflow improves correlation from 0.2 → 0.7+
- [ ] Users can iterate on prompts without code changes

**Phase 6 Success:**
- [ ] API uses validated judges from batch workflow
- [ ] Continuous monitoring detects judge degradation

---

## References

- [initial-spec.md](./initial-spec.md) - Original eval-agent design
- [G-eval Paper](https://arxiv.org/pdf/2303.16634) - Research foundation
