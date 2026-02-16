# KG Agent

Knowledge Graph Agent with Claude for document search and question answering.

## Features

| Feature | Description |
|---------|-------------|
| AWS Bedrock Integration | Claude API for reasoning and embeddings |
| Query Rewriting | Automatic query optimization using Claude |
| Document Ingestion | Parse, chunk, and embed documents |
| Semantic Search | Vector similarity search using pgvector |
| Keyword Search | Full-text search using PostgreSQL |
| Hybrid Search | Combined search with RRF ranking |
| REST API | HTTP endpoints for agent and search |
| RAG Integration | Context-aware responses with search |
| Streaming Responses | Server-Sent Events for real-time output |

## Architecture

```
kg-agent/
├── cmd/
│   ├── agent/       # Main agent service (Claude + orchestration)
│   ├── search/      # Search API service (semantic, keyword, hybrid)
│   └── ingest/      # Document ingestion CLI
├── internal/
│   ├── agent/       # Agent HTTP handlers
│   ├── search/      # Search HTTP handlers and service
│   ├── bedrock/     # AWS Bedrock client
│   ├── embedding/   # Titan embedding service
│   ├── database/    # PostgreSQL operations
│   └── ingestion/   # Document processing pipeline
└── migrations/      # Database schema
```

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- AWS credentials configured
- PostgreSQL with pgvector extension

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

# API Ports
AGENT_API_PORT=8081      # Agent API
SEARCH_API_PORT=8082     # Search API

# Search Client Configuration
SEARCH_API_URL=http://localhost:8082
SEARCH_API_TIMEOUT=15    # seconds
```

## Local Development

### Start PostgreSQL

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
| `/api/v1/query` | POST | Query Claude (non-streaming) |
| `/api/v1/query/stream` | POST | Query Claude (streaming SSE) |

### Search API (Port 8082)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/search/v1/semantic` | POST | Vector similarity search |
| `/search/v1/keyword` | POST | Full-text search |
| `/search/v1/hybrid` | POST | Combined search with RRF |

## Testing

### Agent API (with RAG)

```bash
# Non-streaming query with context
curl -X POST http://localhost:8081/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"prompt": "How do I encrypt my files?", "max_tokens": 500}' | jq .

# Example response:
# {
#   "content": "Based on the documentation, you can encrypt files using...",
#   "stop_reason": "end_turn",
#   "model": "anthropic.claude-3-5-sonnet-20241022-v2:0"
# }

# Streaming query with context
curl -X POST http://localhost:8081/api/v1/query/stream \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Explain two-factor authentication", "max_tokens": 500}'

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
