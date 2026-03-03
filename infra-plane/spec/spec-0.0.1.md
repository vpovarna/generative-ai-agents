 InferenceFlow Gateway - ML Inference Control Plane

     Context

     You need an ML Gateway / Inference Control Plane for routing and managing inference requests across multiple LLM models. The gateway should be production-ready with high throughput, reliability features, and integration with your existing eval-agent for quality monitoring.

     Why this is needed:
     - Current state: No centralized routing for model inference requests
     - Problem: Need intelligent routing, rate limiting, cost optimization, and quality monitoring
     - Goal: Production-ready gateway that routes requests efficiently with reliability features
     - Unique value: Integration with eval-agent for quality-driven routing

     Requirements:
     - Language: Go with channels for high throughput
     - Architecture: Stateless (Redis for shared state)
     - Traffic: Internal only
     - Streaming: Add later (focus on sync initially)
     - Existing integration: eval-agent at localhost:18082

     ---

     What InfraPlane Adds (ML-Specific):

     ┌───────────────────────────┬──────────────────────────────────────────────────────┬──────────────────────────┐
     │          Feature          │                 Why Critical for ML                  │          Impact          │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Token-based rate limiting │ LLM costs are per-token, not per-request             │ Cost control             │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Response caching          │ LLM inference is expensive (~$0.01-0.10 per request) │ 90% cost reduction       │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Model routing             │ Route by cost/latency/capability                     │ Optimize cost vs quality │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Evaluation hooks          │ Real-time quality monitoring with eval-agent         │ Production safety        │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Request coalescing        │ Deduplicate identical concurrent prompts             │ Throughput optimization  │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Fallback chains           │ Model A fails → try Model B → Model C                │ Reliability              │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Priority queues           │ VIP > normal > batch                                 │ SLA enforcement          │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ Cost tracking             │ Track spend per tenant/model                         │ FinOps                   │
     ├───────────────────────────┼──────────────────────────────────────────────────────┼──────────────────────────┤
     │ A/B testing               │ Compare model versions with eval scores              │ Experimentation          │
     └───────────────────────────┴──────────────────────────────────────────────────────┴──────────────────────────┘

     Verdict: Envoy is excellent for generic HTTP routing but lacks ML-specific features. InferenceFlow provides domain-specific capabilities that Envoy can't deliver without complex C++ filters.

     ---
     System Architecture

     ┌─────────────────────────────────────────────────────────────────────┐
     │                        InfraPlane Gateway                         │
     │                                                                       │
     │  ┌──────────────┐    ┌──────────────┐    ┌──────────────────┐      │
     │  │ HTTP Server  │───▶│   Router     │───▶│  Priority Queues │      │
     │  │ (go-restful) │    │  (routing    │    │  - High          │      │
     │  │              │    │   policies)  │    │  - Normal        │      │
     │  └──────────────┘    └──────────────┘    │  - Low           │      │
     │         │                   │             └────────┬─────────┘      │
     │         │                   ▼                      │                 │
     │         │          ┌──────────────────┐            │                 │
     │         │          │ Cache Layer      │            │                 │
     │         │          │ (Redis)          │            │                 │
     │         │          └──────────────────┘            │                 │
     │         │                                           ▼                 │
     │         │                              ┌───────────────────────┐     │
     │         │                              │  Worker Pool          │     │
     │         │                              │  (10 goroutines)      │     │
     │         │                              └───────────┬───────────┘     │
     │         │                                          │                 │
     │         ▼                                          ▼                 │
     │  ┌────────────────────────────────────────────────────────────┐     │
     │  │           Circuit Breaker Manager                          │     │
     │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │     │
     │  │  │ Claude   │  │  GPT-4   │  │  Gemini  │  │  Llama   │  │     │
     │  │  │    CB    │  │    CB    │  │    CB    │  │    CB    │  │     │
     │  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │     │
     │  └────────────────────────────────────────────────────────────┘     │
     │         │              │              │              │               │
     └─────────┼──────────────┼──────────────┼──────────────┼───────────────┘
               │              │              │              │
               ▼              ▼              ▼              ▼
        ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
        │ Bedrock  │   │ OpenAI   │   │  Vertex  │   │  Others  │
        │ Claude   │   │  GPT-4   │   │  Gemini  │   │          │
        └────┬─────┘   └────┬─────┘   └────┬─────┘   └────┬─────┘
             │              │              │              │
             └──────────────┴──────────────┴──────────────┘
                            │
                            ▼
                   ┌────────────────┐
                   │  Eval Agent    │
                   │  (Async Hooks) │
                   │  Port: 18082   │
                   └────────────────┘

     Component Responsibilities

     HTTP Server: Request validation, auth, middleware (logging, CORS, metrics)
     Router: Model selection based on policies (cost/latency/capability)
     Priority Queues: High/normal/low priority channels with select-based dequeue
     Cache Layer: Response cache + request coalescing (Redis-backed)
     Worker Pool: 10 goroutines processing from priority queues
     Circuit Breakers: Per-model fail-fast protection
     Upstream Clients: HTTP connection pools to model endpoints
     Eval Integration: Async quality monitoring (10% sampling)

     ---
     PHASE 1: Core Gateway (MVP)

     Timeline: 4 weeks
     Goal: Working gateway with reliability features

     Features

     ✅ HTTP API with go-restful/v3
     ✅ Single model routing (static config)
     ✅ Worker pool with priority queues (channels)
     ✅ Circuit breaker per model
     ✅ Redis-backed token bucket rate limiting
     ✅ Health checks
     ✅ Structured logging (zerolog)
     ✅ Prometheus metrics
     ✅ YAML configuration
     ✅ Docker build + Makefile

     File Structure

     inferenceflow/
     ├── cmd/
     │   └── gateway/
     │       └── main.go                      # Entry point, server bootstrap
     ├── internal/
     │   ├── api/
     │   │   ├── handler.go                   # HTTP handlers (/v1/infer, /v1/health)
     │   │   ├── routes.go                    # Route registration
     │   │   ├── models.go                    # InferenceRequest/Response structs
     │   │   └── middleware/
     │   │       ├── logging.go               # Request logging
     │   │       ├── recovery.go              # Panic recovery
     │   │       └── metrics.go               # Prometheus middleware
     │   ├── config/
     │   │   ├── config.go                    # Config structs
     │   │   └── loader.go                    # YAML loading + validation
     │   ├── gateway/
     │   │   ├── core.go                      # Gateway coordinator (channels, workers)
     │   │   └── worker.go                    # Worker pool implementation
     │   ├── executor/
     │   │   ├── executor.go                  # Executor interface
     │   │   └── bedrock.go                   # AWS Bedrock implementation
     │   ├── circuitbreaker/
     │   │   ├── breaker.go                   # Circuit breaker state machine
     │   │   └── state.go                     # State transitions
     │   ├── ratelimit/
     │   │   ├── limiter.go                   # Rate limiter interface
     │   │   └── token_bucket.go              # Redis token bucket
     │   ├── router/
     │   │   ├── router.go                    # Model selection logic
     │   │   └── selector.go                  # Routing strategies
     │   ├── metrics/
     │   │   └── prometheus.go                # Metrics registration
     │   └── redis/
     │       └── client.go                    # Redis connection
     ├── configs/
     │   └── gateway.yaml                     # Configuration file
     ├── Dockerfile                           # Multi-stage build
     ├── Makefile                             # Build targets
     ├── go.mod
     └── README.md

    
     Testing Strategy

     Unit Tests:
     - Circuit breaker state transitions (closed→open→half-open)
     - Token bucket logic with mocked Redis
     - Priority queue ordering
     - Request validation

     Integration Tests:
     - Full request flow (HTTP → worker → executor → response)
     - Circuit breaker with simulated failures
     - Rate limiting with real Redis
     - Graceful shutdown

     Load Tests:
     - 1000 RPS sustained load for 10 minutes
     - Queue depth monitoring
     - Circuit breaker activation under failures
     - Rate limit accuracy

     Success Criteria

     Functional:
     - ✅ Routes requests to Bedrock Claude successfully
     - ✅ Circuit breaker opens after 5 consecutive failures
     - ✅ Rate limiter enforces 100 RPM per tenant
     - ✅ Health endpoint reflects circuit breaker states

     Performance:
     - ✅ Handles 1000 RPS with <100ms gateway overhead
     - ✅ Queue depth <500 under normal load
     - ✅ Worker utilization >80%

     Reliability:
     - ✅ Zero dropped requests during graceful shutdown
     - ✅ Circuit breaker prevents cascading failures
     - ✅ Rate limiter <1% false rejections

     ---
     PHASE 2: Production Features

     Timeline: 4 weeks
     Goal: Advanced routing, caching, eval integration, cost tracking

     Features

     ✅ Redis-backed response cache (exact match)
     ✅ Token-aware rate limiting (TPM + RPM)
     ✅ Advanced routing (cost/latency/capability)
     ✅ Fallback chains (Model A→B→C)
     ✅ A/B testing (traffic splitting)
     ✅ Eval-agent integration (async hooks)
     ✅ Request coalescing (deduplicate)
     ✅ Priority queues (enhanced)
     ✅ Cost tracking per tenant
     ✅ Enhanced metrics (per-model, per-tenant)

     Additional Files

     inferenceflow/
     ├── internal/
     │   ├── cache/
     │   │   ├── response_cache.go            # Redis response cache
     │   │   ├── coalescing.go                # Request coalescing
     │   │   └── key_generator.go             # Cache key generation
     │   ├── router/
     │   │   ├── cost_router.go               # Cost-optimized routing
     │   │   ├── latency_router.go            # Latency-optimized routing
     │   │   ├── capability_router.go         # Task-based routing
     │   │   └── ab_test.go                   # A/B testing
     │   ├── fallback/
     │   │   ├── chain.go                     # Fallback chain executor
     │   │   └── strategy.go                  # Fallback strategies
     │   ├── eval/
     │   │   ├── client.go                    # Eval-agent HTTP client
     │   │   ├── sampler.go                   # Sampling logic (10%)
     │   │   └── async_evaluator.go           # Async evaluation
     │   ├── cost/
     │   │   ├── tracker.go                   # Cost tracking (Redis)
     │   │   └── aggregator.go                # Daily/monthly aggregation
     │   └── ratelimit/
     │       └── token_aware.go               # TPM-based limiter
     └── configs/
         └── routing_rules.yaml               # Routing policies, A/B tests

     
     Testing Strategy

     Cache Tests:
     - Verify cache hit/miss accuracy
     - Test key generation consistency
     - Validate TTL expiration

     Routing Tests:
     - Cost router selects cheapest model
     - Fallback chain executes on failure
     - A/B test traffic split accuracy (±2%)

     Eval Integration Tests:
     - Mock eval-agent endpoint
     - Verify 10% sampling rate
     - Test async evaluation (non-blocking)
     - Validate quality score feedback

     Coalescing Tests:
     - Identical requests share result
     - Verify reduction in upstream calls
     - Test timeout handling

     Cost Tracking Tests:
     - Per-tenant cost accuracy (±1%)
     - Daily aggregation correctness
     - Budget alert triggers

     Success Criteria

     Caching:
     - ✅ Cache hit rate >30% for repeated prompts
     - ✅ Coalescing reduces duplicate calls by >50%
     - ✅ Cache latency <10ms

     Routing:
     - ✅ Cost router saves >20% vs always expensive model
     - ✅ Fallback chain success rate >95%
     - ✅ A/B tests split traffic correctly (±2%)

     Evaluation:
     - ✅ 10% of traffic evaluated
     - ✅ Evaluation doesn't block responses
     - ✅ Quality scores update routing decisions

     Cost Tracking:
     - ✅ Per-tenant cost accurate within 1%
     - ✅ Daily aggregation <1s
     - ✅ Budget alerts within 5s

     ---
     Implementation Sequence

     Phase 1 

     Week 1: Foundation
     - Create project structure
     - Config loading (YAML + env vars)
     - Redis connection setup
     - Logging and metrics initialization

     Week 2: Core Gateway
     - Worker pool with priority queues
     - Request routing (simple model selection)
     - Executor interface + Bedrock implementation
     - Channel-based request flow

     Week 3: Reliability
     - Circuit breaker implementation
     - Redis token bucket rate limiter
     - Health check endpoint
     - Graceful shutdown

     Week 4: HTTP API & Testing
     - go-restful handlers (/v1/infer, /v1/health)
     - Middleware (logging, recovery, metrics)
     - Integration tests
     - Load testing (target: 1000 RPS)
     - Documentation

     Phase 2

     Week 5: Caching
     - Response cache (exact match)
     - Request coalescing
     - Cache metrics and monitoring

     Week 6: Advanced Routing
     - Cost-optimized router
     - Latency-optimized router
     - Fallback chains
     - A/B testing infrastructure

     Week 7: Eval Integration
     - Eval-agent HTTP client
     - Async evaluation pipeline
     - Sampling logic (10% of traffic)
     - Quality score feedback to router

     Week 8: Cost & Polish
     - Cost tracking per tenant
     - Token-aware rate limiting (TPM)
     - Enhanced metrics dashboard
     - Load testing (target: 2000 RPS)
     - Production deployment docs