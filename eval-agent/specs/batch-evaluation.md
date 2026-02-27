Batch Dataset Evaluation CLI Implementation Plan                                                                               
                                                                                                                                
 Context                                                                                                                        
                                                                                                                                
 The eval-agent currently supports single-request evaluation via HTTP API, Redis Stream consumer, and MCP protocol. To enable   
 research workflows (prompt optimization, A/B testing, correlation analysis), we need a batch evaluation CLI that can process
 datasets of evaluation requests efficiently.

 Why this is needed:
 - Enable evaluation of large test datasets (20-1000+ cases)
 - Support research workflows: prompt iteration, judge comparison, metrics analysis
 - Allow offline evaluation without running API server
 - Generate reports for dataset quality assessment

 Current state:
 - ✅ YAML-driven judges (Phase 1 completed)
 - ✅ Single evaluation pipeline (Executor.Execute)
 - ✅ Test dataset exists (docs/test-dataset-relevance.json with 20 cases)
 - ❌ No batch processing capability

 Target state:
 - Read JSONL dataset (one evaluation request per line)
 - Process concurrently with configurable worker pool
 - Output results in multiple formats (JSONL, CSV, summary)
 - Show progress and handle errors gracefully

 ---
 Architecture Overview

 Input/Output Flow

 JSONL File → Reader (streaming) → Job Channel → Worker Pool → Output Channel → Writer (JSONL/CSV/summary)
                                       ↓
                               Executor.Execute()
                               (reuses existing pipeline)

 Components (All New - No Changes to Existing Code)

 New files to create:
 - cmd/batch/main.go - CLI entry point with flag parsing
 - internal/batch/reader.go - JSONL streaming reader
 - internal/batch/processor.go - Worker pool coordinator
 - internal/batch/progress.go - Thread-safe progress tracker
 - internal/batch/writer.go - Output formatters (JSONL/CSV/summary)
 - internal/batch/types.go - Batch-specific types

 Integration points (reuse existing):
 - setup.Wire() - Dependency injection
 - executor.Executor.Execute() - Evaluation logic
 - models.EvaluationRequest/Result - Data structures
 - configs/judges.yaml - Judge configuration

 ---
 CLI Interface

 Command Signature

 go run cmd/batch/main.go [flags]

 Flags

 ┌────────────────────┬────────┬──────────┬──────────────────────────────────────────┐
 │        Flag        │  Type  │ Default  │               Description                │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -input             │ string │ required │ Input JSONL file path (or "-" for stdin) │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -output            │ string │ stdout   │ Output file path                         │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -format            │ string │ "jsonl"  │ Output format: "jsonl", "csv", "summary" │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -summary           │ string │ ""       │ Optional separate summary file           │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -workers           │ int    │ 5        │ Concurrent evaluation workers            │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -continue-on-error │ bool   │ true     │ Continue on evaluation failures          │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -dry-run           │ bool   │ false    │ Validate input without evaluating        │
 ├────────────────────┼────────┼──────────┼──────────────────────────────────────────┤
 │ -progress-interval │ int    │ 5        │ Progress log interval (seconds)          │
 └────────────────────┴────────┴──────────┴──────────────────────────────────────────┘

 Usage Examples

 # Basic batch evaluation
 go run cmd/batch/main.go -input dataset.jsonl -output results.jsonl

 # High concurrency for large datasets
 go run cmd/batch/main.go -input dataset.jsonl -workers 10

 # CSV export for spreadsheet analysis
 go run cmd/batch/main.go -input dataset.jsonl -format csv -output results.csv

 # Summary report only
 go run cmd/batch/main.go -input dataset.jsonl -format summary

 # Combined: results + summary
 go run cmd/batch/main.go -input dataset.jsonl -output results.jsonl -summary summary.json

 # Pipeline from stdin
 cat dataset.jsonl | go run cmd/batch/main.go -input - | jq 'select(.verdict=="fail")'

 ---
 Implementation Details

 1. JSONL Reader (internal/batch/reader.go)

 Responsibility: Stream JSONL input line-by-line, parse into EvaluationRequest

 Key features:
 - Uses bufio.Scanner for memory-efficient streaming
 - Returns channel of InputRecord (line number + request + error)
 - Skips empty lines
 - Captures parse errors with line numbers
 - Context-cancellable

 Pattern:
 type InputRecord struct {
     LineNumber int
     Request    models.EvaluationRequest
     Error      error
 }

 reader := batch.NewReader(file)
 records := reader.ReadAll(ctx) // Returns <-chan InputRecord

 2. Batch Processor (internal/batch/processor.go)

 Responsibility: Coordinate worker pool, distribute evaluations, collect results

 Key features:
 - Worker pool pattern (similar to judge runner)
 - Fixed number of workers (default 5, configurable)
 - Uses sync.WaitGroup for coordination
 - Context propagation for cancellation
 - Per-evaluation timing

 Pattern:
 processor := batch.NewProcessor(executor, workers, logger)
 outputs := processor.Process(ctx, jobChannel) // Returns <-chan EvaluationOutput

 Worker implementation:
 - Converts EvaluationRequest → EvaluationContext
 - Calls executor.Execute(ctx, evalCtx)
 - Captures timing and errors
 - Sends to output channel

 3. Progress Tracker (internal/batch/progress.go)

 Responsibility: Thread-safe progress tracking with periodic logging

 Key features:
 - Atomic counters (total, processed, succeeded, failed, skipped)
 - Ticker-based periodic logging (follows existing patterns)
 - Non-blocking updates
 - Final stats collection

 Pattern:
 tracker := batch.NewProgressTracker(total, interval, logger)
 tracker.RecordSuccess() // Atomic increment
 tracker.RecordFailure() // Atomic increment
 stats := tracker.GetStats() // Final summary

 4. Output Writers (internal/batch/writer.go)

 Responsibility: Format and write results

 Formats supported:

 A. JSONL Writer
 - One EvaluationResult JSON per line
 - Directly pipeable to jq, Python pandas
 - Preserves all evaluation details

 B. CSV Writer
 - Header row: id, verdict, confidence, stage scores
 - Columns for each precheck and judge score
 - Easy import to Excel/Google Sheets

 C. Summary Writer
 - Aggregate statistics: total, succeeded, failed
 - Verdict distribution: pass/fail/review counts
 - Average confidence, average duration
 - Error list with line numbers

 5. Main CLI (cmd/batch/main.go)

 Workflow:
 1. Parse flags, validate required arguments
 2. Load environment (.env file)
 3. Setup logging (zerolog console writer)
 4. Open input (file or stdin) and output (file or stdout)
 5. Wire dependencies with setup.Wire()
 6. Stream input with Reader
 7. Collect jobs, count total for progress tracking
 8. If dry-run, exit after validation
 9. Create ProgressTracker
 10. Start worker pool with Processor
 11. Write results with appropriate Writer
 12. Collect statistics (verdicts, confidence, timing)
 13. Write summary if requested
 14. Log final stats and exit

 Error handling:
 - Input errors (malformed JSON): log, skip, continue
 - Evaluation errors (LLM failures): log, record, continue/exit based on flag
 - Fatal errors (file I/O, setup): log, exit immediately

 Graceful shutdown:
 - Handle SIGINT/SIGTERM signals
 - Cancel context to stop workers
 - Wait for in-flight evaluations
 - Write partial results
