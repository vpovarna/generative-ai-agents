# KG Agent

Knowledge Graph Agent with Claude for document search and question answering.

## Features

| Feature | Description |
|---------|-------------|
| AWS Bedrock Integration | Claude API for reasoning and embeddings |
| Query Rewriting | Automatic query optimization using Claude |
| Retrieval Strategy | Smart decision: search vs. answer from memory |
| Document Ingestion | Parse, chunk, and embed documents |
| Semantic Search | Vector similarity search using pgvector |
| Keyword Search | Full-text search using PostgreSQL |
| Hybrid Search | Combined search with RRF ranking |
| REST API | HTTP endpoints for agent and search |
| RAG Integration | Context-aware responses with search |
| Streaming Responses | Server-Sent Events for real-time output |
| Conversation Memory | Redis-backed multi-turn conversations |
| Session Management | Automatic session creation and tracking |

## Architecture

```
kg-agent/
├── cmd/
│   ├── agent/       # Main agent service (Claude + orchestration)
│   ├── search/      # Search API service (semantic, keyword, hybrid)
│   └── ingest/      # Document ingestion CLI
├── internal/
│   ├── agent/       # Agent HTTP handlers and service
│   ├── search/      # Search HTTP handlers and service
│   ├── strategy/    # Retrieval strategy (heuristic + LLM classifier)
│   ├── conversation/# Conversation memory (Redis)
│   ├── rewrite/     # Query rewriting service
│   ├── bedrock/     # AWS Bedrock client
│   ├── embedding/   # Titan embedding service
│   ├── database/    # PostgreSQL operations
│   ├── redis/       # Redis connection
│   ├── middleware/  # HTTP middleware (logging, errors)
│   └── ingestion/   # Document processing pipeline
└── migrations/      # Database schema
```

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- AWS credentials configured
- PostgreSQL with pgvector extension
- Redis 7+

## Environment Setup

Create a `.env` file:

```bash
# AWS Configuration
AWS_REGION=us-east-1
CLAUDE_MODEL_ID=anthropic.claude-3-5-sonnet-20241022-v2:0

# Database Configuration
KG_AGENT_VECTOR_DB_HOST=localhost
KG_AGENT_VECTOR_DB_PORT=5432
KG_AGENT_VECTOR_DB_USER=postgres
KG_AGENT_VECTOR_DB_PASSWORD=postgres
KG_AGENT_VECTOR_DB_DATABASE=kg_agent
KG_AGENT_VECTOR_DB_SSLMode=disable

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_TTL=30m              # Conversation TTL

# API Ports
AGENT_API_PORT=8081         # Agent API
SEARCH_API_PORT=8082        # Search API

# Search Client Configuration
SEARCH_API_URL=http://localhost:8082
SEARCH_API_TIMEOUT=15       # seconds
SEARCH_API_MAX_IDLE_CONNS=100
SEARCH_API_MAX_IDLE_CONNS_PER_HOST=10
```

## Local Development

### Start PostgreSQL and Redis

```bash
docker-compose up -d
```

### Ingest Documents

```bash
# Ingest a single document
go run cmd/ingest/main.go -insert-doc -filePath resources/product-guide.txt

# With custom chunking parameters
go run cmd/ingest/main.go -insert-doc -filePath resources/api-docs.txt -chunkSize 1000 -chunkOverlap 200

# Get all ingested documents
go run cmd/ingest/main.go -get-docs

# Delete by doc ID
go run cmd/ingest/main.go -delete-doc  -doc-id 9123a9ca-073d-449d-a5e3-024d54e1e15c
```

### Start Services

```bash
# Start Search API (port 8082)
go run cmd/search/main.go

# Start Agent API (port 8081)
go run cmd/agent/main.go
```

## API Endpoints

### Agent API (Port 8081)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/health` | GET | Health check |
| `/api/v1/query` | POST | Query Claude (non-streaming, with conversation memory) |
| `/api/v1/query/stream` | POST | Query Claude (streaming SSE, with conversation memory) |

### Search API (Port 8082)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/search/v1/semantic` | POST | Vector similarity search |
| `/search/v1/keyword` | POST | Full-text search |
| `/search/v1/hybrid` | POST | Combined search with RRF |

## Testing

### Test Retrieval Strategy (Smart Search Decision)

The agent now intelligently decides when to search documentation vs. answer from conversation history.

**Test 1: Greeting (No Search)**
```bash
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "max_tokens": 100}' | jq .

# Expected: Agent responds without searching (greeting detected by heuristic)
```

**Test 2: New Technical Question (Searches)**
```bash
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"prompt": "How do I encrypt my files?", "max_tokens": 500}' | jq .

# Save the session_id from response!
# Expected: Agent searches documentation and provides detailed answer
```

**Test 3: Follow-up Question (No Search)**
```bash
# Use session_id from Test 2
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "YOUR_SESSION_ID_HERE",
    "prompt": "tell me more about that",
    "max_tokens": 500
  }' | jq .

# Expected: Agent answers from conversation history without searching
# (Follow-up detected by heuristic)
```

**Test 4: Pronoun Reference (No Search)**
```bash
# Use same session_id
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "YOUR_SESSION_ID_HERE",
    "prompt": "What are the performance implications of it?",
    "max_tokens": 500
  }' | jq .

# Expected: Agent resolves "it" from context without searching
```

**Test 5: New Topic with History (Searches)**
```bash
# Use same session_id, but new unrelated topic
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "YOUR_SESSION_ID_HERE",
    "prompt": "How do I configure SSL certificates?",
    "max_tokens": 500
  }' | jq .

# Expected: Agent searches (new topic not in history)
```

**Test 6: Complex Ambiguous Query (LLM Classifier)**
```bash
# Start new session - ambiguous query
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"prompt": "What about version 2 differences?", "max_tokens": 500}' | jq .

# Expected: Heuristic has low confidence, LLM classifier decides
# (Should search since no context for "version 2")
```

### Monitor Decision Logic

Watch the logs to see retrieval decisions:

```bash
# In agent terminal, you'll see:
# {"level":"info","method":"heuristic","message":"Using heuristic decision"}
# OR
# {"level":"info","method":"llm_fallback","message":"Low confidence, using LLM classifier"}
```

### Agent API (with RAG)

### Conversation Memory Testing

```bash
# First query - creates new session
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"prompt": "How do I encrypt my files?", "max_tokens": 500}' | jq .

# Example response:
# {
#   "session_id": "550e8400-e29b-41d4-a716-446655440000",
#   "content": "Based on the documentation, you can encrypt files using...",
#   "stop_reason": "end_turn",
#   "model": "anthropic.claude-3-5-sonnet-20241022-v2:0"
# }

# Follow-up query - uses existing session (copy session_id from above)
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "prompt": "What about performance impact?",
    "max_tokens": 500
  }' | jq .

# Streaming query with conversation memory
curl -N -X POST http://localhost:8081/api/v1/query/stream \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "prompt": "Can you summarize what we discussed?",
    "max_tokens": 500
  }'

# Health check
curl http://localhost:8081/api/v1/health | jq .
```

### Search API

```bash
# Semantic search (vector similarity)
curl -X POST http://localhost:8082/search/v1/semantic \
  -H "Content-Type: application/json" \
  -d '{"query": "How do I secure my files?", "limit": 3}' | jq .

# Keyword search (full-text)
curl -X POST http://localhost:8082/search/v1/keyword \
  -H "Content-Type: application/json" \
  -d '{"query": "encryption AES-256", "limit": 3}' | jq .

# Hybrid search (combined with RRF ranking)
curl -X POST http://localhost:8082/search/v1/hybrid \
  -H "Content-Type: application/json" \
  -d '{"query": "two-factor authentication setup", "limit": 5}' | jq .

# Example hybrid search response:
# {
#   "query": "two-factor authentication setup",
#   "result": [
#     {
#       "chunk_id": "abc-123",
#       "document_id": "doc-456",
#       "content": "To enable 2FA, go to Settings...",
#       "score": 0.85,
#       "rank": 1
#     }
#   ],
#   "count": 5,
#   "method": "hybrid"
# }
```

### Debug Redis Conversations

```bash
# Connect to Redis
docker exec -it kg-agent-redis-1 redis-cli

# List all active sessions
SMEMBERS active_sessions

# Get conversation for a specific session
GET conversation:550e8400-e29b-41d4-a716-446655440000

# Check TTL (time to live) for a session
TTL conversation:550e8400-e29b-41d4-a716-446655440000

# Delete a session manually (for testing)
DEL conversation:550e8400-e29b-41d4-a716-446655440000
SREM active_sessions 550e8400-e29b-41d4-a716-446655440000
```

## Development Commands

### Build

```bash
# Build all services
go build -o bin/agent cmd/agent/main.go
go build -o bin/search cmd/search/main.go
go build -o bin/ingest cmd/ingest/main.go
```

## License

MIT
