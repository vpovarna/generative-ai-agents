# Real-Time Evaluation Agent – Spec (MVP)

## Goal

Build a minimal evaluation service that:
- Consumes agent responses (query, context, answer)
- Uses LLM-as-Judge to score them
- Computes a structured confidence score

Goal: Eval agent using streaming (Kafka or any other streaming solution), backpressure. The agent should allow adding other metrics for improving the confidence score. 

---

## Objective
Create a standalone evaluation engine that scores a single agent response using LLM-as-judge.

## User flow:

 - Read Kafka Message  
 - Normalize to EvaluationContext  
 - Build Execution Plan  
 - Execute Stage 1: parallel fast checks  
 - Early exit? (if very bad)  
 - Execute Stage 2 (parallel LLM judge)  
 - Aggregate  
 - Emit EvaluationResult  
---

## Execution Plan
Step 1 — Core Types

Define the data contracts everything else builds on:
 - EvaluationRequest — {query, context, answer, ...}
 - EvaluationContext — normalized internal form
 - StageResult — {name, score float64, reason string, duration}
 - EvaluationResult — {id, stages []StageResult, confidence float64, verdict string}

Step 2 — Stage 1: Fast Checks (no LLM)

Cheap heuristics that run in parallel, cheap to fail fast: 
 - Length check — answer too short/long relative to query
 - Overlap check — keyword overlap between query and answer
 - Format check — answer is non-empty, not malformed

Step 3 — Stage 2: LLM-as-Judge (parallel)

Three parallel LLM calls, each using a focused prompt:
 - Faithfulness — does the answer follow from the context?
 - Relevance — does the answer address the query?
 - Coherence — is the answer logically consistent?

Step 4 — Aggregator

Takes []StageResult from both stages, computes a weighted confidence float64, produces the final EvaluationResult. Weights are configurable to make the scoring strategy swappable.

Step 5 — Kafka Consumer with Backpressure

 - Consumer reads messages, decodes to EvaluationRequest
 - Worker pool (bounded via buffered channel or semaphore) processes evaluations
 - Manual offset commit after successful evaluation (at-least-once)
 - Producer writes EvaluationResult back to output topic

Step 6 — Config & Wiring

 - config.go — load from env (Kafka brokers, topics, LLM key, thresholds, weights, concurrency)