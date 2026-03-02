# Search Service

A high-performance vector search service built in Go, providing semantic, keyword, and hybrid search capabilities over document embeddings stored in PostgreSQL with pgvector.

## Features

- **Semantic Search**: Vector similarity search using AWS Bedrock embeddings (cosine distance)
- **Keyword Search**: Full-text search using PostgreSQL's native text search capabilities
- **Hybrid Search**: Combined approach using Reciprocal Rank Fusion (RRF) to merge semantic and keyword results
- **RESTful API**: Clean HTTP API with JSON request/response
- **AWS Bedrock Integration**: Generates embeddings using Amazon Titan or other Bedrock embedding models
- **PostgreSQL + pgvector**: Efficient vector storage and similarity search

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│     Search Service (Go)         │
│  ┌───────────────────────────┐  │
│  │   HTTP API Layer          │  │
│  │  - Semantic Search        │  │
│  │  - Keyword Search         │  │
│  │  - Hybrid Search (RRF)    │  │
│  └───────────┬───────────────┘  │
│              │                   │
│  ┌───────────▼───────────────┐  │
│  │   Search Service Logic    │  │
│  │  - RRF Algorithm          │  │
│  │  - Score Normalization    │  │
│  └───────────┬───────────────┘  │
│              │                   │
│  ┌───────────▼───────────────┐  │
│  │   AWS Bedrock Client      │  │
│  │  - Embedding Generation   │  │
│  └───────────────────────────┘  │
└──────────────┬──────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│  PostgreSQL + pgvector           │
│  - Document chunks               │
│  - Vector embeddings             │
│  - Full-text search indexes      │
└──────────────────────────────────┘
```

## API Endpoints

Base path: `/search/v1`

### POST `/search/v1/semantic`
Performs vector similarity search using embeddings.

**Request:**
```json
{
  "query": "What is machine learning?",
  "limit": 10
}
```

**Response:**
```json
{
  "query": "What is machine learning?",
  "result": [
    {
      "chunk_id": "uuid",
      "document_id": "doc-uuid",
      "content": "Machine learning is...",
      "score": 0.92,
      "rank": 1,
      "metadata": {
        "source": "ml-guide.pdf",
        "page": "1"
      }
    }
  ],
  "count": 10,
  "method": "semantic"
}
```

### POST `/search/v1/keyword`
Performs full-text keyword search using PostgreSQL's text search.

**Request:**
```json
{
  "query": "machine learning",
  "limit": 10
}
```

**Response:** Same format as semantic search, with `"method": "keyword"`.

### POST `/search/v1/hybrid`
Performs hybrid search combining semantic and keyword results using Reciprocal Rank Fusion.

**Request:**
```json
{
  "query": "machine learning applications",
  "limit": 10
}
```

**Response:** Same format, with `"method": "hybrid"`.

**How RRF Works:**
1. Fetch 2×limit results from both semantic and keyword search
2. Calculate RRF score for each result: `score = 1 / (k + rank)` where k=60
3. If a chunk appears in both results, sum the RRF scores
4. Sort by combined score and return top N results

## Prerequisites

- Go 1.25.6 or higher
- PostgreSQL 15+ with pgvector extension
- AWS account with Bedrock access
- AWS credentials configured

## Setup

### 1. Clone and Install Dependencies

```bash
cd search-service
go mod download
```

### 2. Database Setup

Run the migrations to set up PostgreSQL with pgvector:

```bash
psql -U your_user -d your_database -f migrations/001_enable_extensions.sql
psql -U your_user -d your_database -f migrations/002_create_documents_table.sql
psql -U your_user -d your_database -f migrations/003_create_chunks_table.sql
```

### 3. Environment Configuration

Create a `.env` file in the root directory:

```bash
# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
BEDROCK_ENDPOINT=https://bedrock-runtime.us-east-1.amazonaws.com

# Database Configuration
KG_AGENT_VECTOR_DB_HOST=localhost
KG_AGENT_VECTOR_DB_PORT=5432
KG_AGENT_VECTOR_DB_USER=your_user
KG_AGENT_VECTOR_DB_PASSWORD=your_password
KG_AGENT_VECTOR_DB_DATABASE=your_database
KG_AGENT_VECTOR_DB_SSLMode=disable

# API Configuration
SEARCH_API_PORT=8082
```

## Running the Service

### Build and Run

```bash
# Build
go build -o search-service ./cmd/search

# Run
./search-service
```

### Development Mode

```bash
go run ./cmd/search/main.go
```

The service will start on `http://localhost:8082` (or the port specified in `.env`).

## Usage Examples

### Semantic Search
```bash
curl -X POST http://localhost:8082/search/v1/semantic \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What are the benefits of cloud computing?",
    "limit": 5
  }'
```

### Keyword Search
```bash
curl -X POST http://localhost:8082/search/v1/keyword \
  -H "Content-Type: application/json" \
  -d '{
    "query": "cloud computing benefits",
    "limit": 5
  }'
```

### Hybrid Search
```bash
curl -X POST http://localhost:8082/search/v1/hybrid \
  -H "Content-Type: application/json" \
  -d '{
    "query": "cloud computing advantages",
    "limit": 5
  }'
```

## Project Structure

```
search-service/
├── cmd/
│   ├── search/          # Search API server
│   └── ingest/          # Document ingestion (optional)
├── internal/
│   ├── search/          # Search logic and handlers
│   ├── database/        # PostgreSQL + pgvector integration
│   ├── embedding/       # AWS Bedrock embedding client
│   ├── bedrock/         # AWS Bedrock client
│   ├── ingestion/       # Document parsing and chunking
│   └── middleware/      # HTTP middleware (logging, errors)
├── migrations/          # Database schema migrations
├── configs/             # Configuration files
├── go.mod
└── README.md
```

## Performance

- **Concurrent Requests**: Go's goroutines handle multiple concurrent search requests efficiently
- **Connection Pooling**: PostgreSQL connection pool managed by pgx
- **Caching**: Consider adding Redis caching for frequently searched queries
- **Batch Processing**: Ingest command supports batch document processing

## Search Strategies

### When to Use Each Method

- **Semantic Search**: Best for conceptual queries, synonyms, and semantic similarity
  - Example: "What is AI?" → matches "artificial intelligence", "machine learning"

- **Keyword Search**: Best for exact matches, proper nouns, and specific terms
  - Example: "Docker container" → matches exact occurrences

- **Hybrid Search**: Best for general-purpose search combining both approaches
  - Example: "Python machine learning libraries" → balances semantic understanding with keyword precision

## Development

### Adding New Search Methods

1. Add method to `internal/search/service.go`
2. Add handler in `internal/search/handler.go`
3. Register route in `internal/search/routes.go`

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/search
```

## Troubleshooting

### Connection Issues
- Verify PostgreSQL is running and accessible
- Check database credentials in `.env`
- Ensure pgvector extension is installed: `CREATE EXTENSION vector;`

### AWS Bedrock Issues
- Verify AWS credentials are configured correctly
- Check Bedrock model availability in your region
- Ensure IAM permissions include `bedrock:InvokeModel`

### Search Returns No Results
- Verify data has been ingested (use `cmd/ingest`)
- Check that embeddings were generated correctly
- Verify vector dimensions match between ingestion and search

## License

See [LICENSE](../LICENSE) file in repository root.

## Contributing

This is part of the generative-ai-agents repository. See main README for contribution guidelines.
