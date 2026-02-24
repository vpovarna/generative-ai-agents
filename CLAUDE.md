# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A learning repository focused on building Generative AI applications with Go, featuring:
- **go-fundamentals**: LeetCode and Advent of Code solutions for Go practice
- **kg-agent**: Production-ready RAG agent with AWS Bedrock Claude, semantic search, conversation memory, guardrails, and caching
- **eval-agent**: LLM-as-Judge evaluation service with heuristic prechecks and parallel judge scoring
- **streamlit-ui**: Python/Streamlit frontend for the agents

## Common Commands

### Testing
```bash
# Run all tests across all modules
make test

# Test a specific module
cd kg-agent && go test -v -race ./...
cd eval-agent && go test -v -race ./...
cd go-fundamentals && go test -v ./...

# Test a specific package
go test -v ./internal/agent/...

# Test with coverage
go test -cover ./...
```

### Linting
```bash
# Lint all code (requires golangci-lint installed)
make lint

# Lint specific module
cd kg-agent && golangci-lint run ./...
```

### Running Services

**KG Agent** (Knowledge Graph RAG Agent):
```bash
cd kg-agent

# Start infrastructure (PostgreSQL + Redis)
docker-compose up -d

# Ingest documents into vector DB
go run cmd/ingest/main.go -insert-doc -filePath resources/product-guide.txt

# Run Search API (port 8082)
go run cmd/search/main.go

# Run Agent API (port 8081)
go run cmd/agent/main.go
```

**Eval Agent** (LLM-as-Judge evaluation):
```bash
cd eval-agent

# Run HTTP API mode (port 18081)
go run cmd/api/main.go

# Run Redis Stream consumer mode
go run cmd/main.go

# Send test evaluation via producer CLI
go run cmd/producer/main.go -d '{"event_id":"evt-001","event_type":"agent_response",...}'
```

### Development Tools
```bash
# Install all Go tools (gopls, delve, air, etc.)
make install-tools

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy

# Clean build artifacts
make clean
```

## Architecture Patterns

### Project Structure
Both production services (kg-agent, eval-agent) follow standard Go layout:
- **cmd/**: Entry points for different binaries (agent, search, ingest, producer, etc.)
- **internal/**: Private application code, organized by domain
- **migrations/**: Database schema migrations (PostgreSQL)
- **resources/**: Sample documents, payloads, test data
- **.env**: Environment configuration (never commit this)

### Common Architectural Patterns

**Dependency Injection**: Services are composed in main.go with explicit dependencies:
```go
service := agent.NewService(
    bedrockClient,     // AWS Bedrock Claude client
    miniClient,        // Haiku for fast operations
    modelID,
    rewriter,          // Query rewriting
    searchClient,      // HTTP client to search API
    conversationStore, // Redis-backed memory
    retrievalStrategy, // Decides when to search
    searchCache,       // Redis cache for search results
)
```

**Middleware Pattern**: Both services use go-restful filters for cross-cutting concerns:
- `middleware.Logger` - structured logging with zerolog
- `middleware.RecoverPanic` - panic recovery and error handling

**Multi-Stage Pipelines**: eval-agent uses a two-stage evaluation:
1. Stage 1: Fast heuristic checks (no LLM) with early-exit on low scores
2. Stage 2: Parallel LLM judge calls (relevance, faithfulness, coherence)
3. Weighted aggregation: (stage1 × 0.3) + (stage2 × 0.7)

**Redis Streams**: eval-agent supports async event processing via Redis Streams with consumer groups

### KG Agent Request Flow

1. **Guardrails** (two layers):
   - Static ban-word check (~1ms) → 400 on match
   - Claude Haiku validation (~500ms) → checks toxicity, PII, prompt injection, off-topic

2. **Conversation Memory**: Load session from Redis or create new session

3. **Retrieval Strategy**: Heuristic + LLM fallback decides whether to search docs or answer from memory
   - Greetings/follow-ups/pronouns → skip search
   - New technical questions → trigger search

4. **Query Rewriting**: Claude Haiku rewrites query for better retrieval

5. **Search Cache**: SHA256-keyed Redis cache (30min TTL) → 10x faster on cache hit

6. **Search**: Hybrid search combining semantic (pgvector cosine) + keyword (PostgreSQL tsvector) with RRF ranking

7. **Model Selection**: Haiku for simple queries, Sonnet for complex queries with context

8. **Streaming**: SSE for real-time token streaming

### Eval Agent Pipeline

**Stage 1 - PreChecks** (parallel, no LLM):
- LengthChecker: answer/query length ratio
- OverlapChecker: keyword overlap between query and answer
- FormatChecker: non-empty, minimum words, no repeated punctuation
- Early exit if avg score < threshold (default 0.2) → skip expensive LLM calls

**Stage 2 - LLM Judges** (parallel Claude calls):
- RelevanceJudge: does answer address the query?
- FaithfulnessJudge: is answer grounded in context (no hallucinations)?
- CoherenceJudge: is answer internally consistent?

**Output**: Confidence score (0.0-1.0) and verdict (pass/review/fail)

## Key Dependencies

- **AWS Bedrock**: Claude API for reasoning, embeddings, and validation
- **PostgreSQL + pgvector**: Vector database for semantic search
- **Redis**: Conversation memory, search caching, and stream processing
- **go-restful**: REST API framework with OpenAPI support
- **zerolog**: Structured logging
- **pgx/v5**: PostgreSQL driver

## Environment Variables

Each service requires a `.env` file in its directory.

**KG Agent**:
- `AWS_REGION`, `CLAUDE_MODEL_ID`, `CLAUDE_MINI_MODEL_ID`
- `KG_AGENT_VECTOR_DB_*`: PostgreSQL connection
- `REDIS_ADDR`, `REDIS_TTL`
- `AGENT_API_PORT` (8081), `SEARCH_API_PORT` (8082)
- `SEARCH_API_URL`, `SEARCH_API_TIMEOUT`

**Eval Agent**:
- `AWS_REGION`, `CLAUDE_MODEL_ID`
- `EVAL_AGENT_API_PORT` (18081)
- `EARLY_EXIT_THRESHOLD` (0.2)
- `REDIS_ADDR`, `REDIS_PASSWORD` (for stream mode)

## Testing Endpoints

**KG Agent**:
```bash
# Query with conversation memory
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"prompt": "How do I encrypt files?", "max_tokens": 500}'

# Streaming query (SSE)
curl -N -X POST http://localhost:8081/api/v1/query/stream \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Explain encryption", "max_tokens": 500}'

# Clear search cache
curl -X POST http://localhost:8081/api/v1/admin/cache/clear
```

**Eval Agent**:
```bash
# Evaluate an agent response
curl -X POST http://localhost:18081/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{"event_id":"evt-001","interaction":{"user_query":"...","context":"...","answer":"..."}}'
```

## Database

**Migrations**: kg-agent uses SQL migrations in `kg-agent/migrations/` that auto-run on PostgreSQL container startup via docker-compose volume mount:
1. `001_enable_extensions.sql` - enables pgvector extension
2. `002_create_documents_table.sql` - stores ingested documents
3. `003_create_chunks_table.sql` - stores document chunks with embeddings

**Running migrations manually**: Migrations run automatically, but you can also use `golang-migrate/migrate` CLI if needed.

## Code Organization

### Internal Package Patterns

Both kg-agent and eval-agent follow domain-driven internal structure:

**kg-agent/internal/**:
- `agent/` - Main query handler, service, routes
- `bedrock/` - AWS Bedrock client wrapper for Claude
- `cache/` - Redis search cache implementation
- `conversation/` - Redis-backed conversation memory
- `database/` - PostgreSQL repository for documents/chunks
- `embedding/` - Bedrock embeddings client
- `guardrails/` - Two-layer input validation (static + Claude)
- `ingestion/` - Document parsing, chunking, embedding pipeline
- `middleware/` - Logging and error recovery filters
- `rewrite/` - Query rewriting with Claude
- `search/` - Semantic, keyword, and hybrid search
- `strategy/` - Retrieval strategy decision logic

**eval-agent/internal/**:
- `api/` - HTTP API handlers and routes
- `aggregator/` - Weighted score aggregation
- `bedrock/` - AWS Bedrock client for judges
- `executor/` - Orchestrates precheck + judge pipeline
- `judge/` - LLM judges (relevance, faithfulness, coherence)
- `prechecks/` - Heuristic checks (length, overlap, format)
- `stream/` - Redis Stream consumer
- `models/` - Shared type definitions

### Interface-Driven Design

Services use interfaces for testability and swappability:
- `conversation.ConversationStore` - abstract conversation storage
- `cache.SearchCache` - abstract search caching
- `prechecks.Checker` - uniform precheck interface
- `judge.Judge` - uniform LLM judge interface

## Redis Integration

**KG Agent uses Redis for**:
- Conversation history (session_id → message array)
- Search result caching (SHA256 query hash → results, 30min TTL)

**Eval Agent uses Redis for**:
- Stream-based async evaluation (consumer group pattern)
- Stream: `eval-events`, Group: `eval-group`

**Useful Redis commands**:
```bash
# Connect to Redis container
docker exec -it kg-agent-redis-1 redis-cli

# List cache keys
KEYS "search_cache:*"

# View stream entries
XRANGE eval-events - + COUNT 10
```

## Working with Go Modules

Each project is a separate Go module:
- `go-fundamentals/go.mod` - minimal, for practice problems
- `kg-agent/go.mod` - includes AWS SDK, pgx, Redis, go-restful
- `eval-agent/go.mod` - includes AWS SDK, Redis, go-restful

When adding dependencies to a service:
```bash
cd kg-agent  # or eval-agent
go get github.com/some/package
go mod tidy
```

## Bedrock Integration

Both services use AWS Bedrock for Claude access:
- Requires AWS credentials configured (`aws configure` or env vars)
- Uses Claude 3.5 Sonnet for complex reasoning
- Uses Claude 3 Haiku for fast operations (validation, classification, rewriting)
- Streaming support via AWS SDK eventstream protocol

**Models**:
- Sonnet: `anthropic.claude-3-5-sonnet-20241022-v2:0`
- Haiku: `anthropic.claude-3-haiku-20240307-v1:0`

## Common Debugging

**Check service health**:
```bash
curl http://localhost:8081/api/v1/health  # kg-agent
curl http://localhost:18081/api/v1/health # eval-agent
```

**Check Docker services**:
```bash
cd kg-agent
docker-compose ps
docker-compose logs -f vector-db
docker-compose logs -f redis
```

**View Redis conversation data**:
```bash
docker exec -it kg-agent-redis-1 redis-cli
GET "conversation:<session_id>"
```

**Check PostgreSQL data**:
```bash
docker exec -it kg-agent-vector-db-1 psql -U postgres -d kg_agent
\dt  # list tables
SELECT COUNT(*) FROM chunks;
```
