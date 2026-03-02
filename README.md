# Generative AI Agents

A collection of AI agents and services built with Go, featuring RAG, evaluation, and search capabilities powered by AWS Bedrock and OpenAI.

## Services

### [eval-agent](./eval-agent)
LLM-as-Judge evaluation service that automatically scores AI agent responses across 5 dimensions (relevance, faithfulness, coherence, completeness, instruction-following). Features include heuristic prechecks, parallel judge execution, batch processing, HTTP API, streaming consumer, and MCP server support.

**Key Features:** Two-stage pipeline (fast heuristics → LLM judges), concurrent worker pools, YAML-configurable judges, JSONL batch processing, streaming support

---

### [kg-agent](./kg-agent)
A RAG agent with AWS Bedrock Claude integration. Provides intelligent question-answering over documents with conversation memory, adaptive model selection, query rewriting, and two-layer guardrails.

**Key Features:** Smart retrieval strategy, Redis-backed conversation memory, search result caching, streaming responses, ban-word + LLM safety validation

---

### [search-service](./search-service)
High-performance vector search service providing semantic, keyword, and hybrid search over document embeddings. Built on PostgreSQL with pgvector extension and AWS Bedrock embeddings.

**Key Features:** Semantic search (cosine similarity), keyword search (full-text), hybrid search (RRF fusion), document ingestion pipeline, RESTful API

---

### [streamlit-ui](./streamlit-ui)
Python/Streamlit frontend for interacting with the agents. Provides a web-based chat interface with conversation history and streaming support.

**Key Features:** Chat interface, session management, streaming responses, integration with kg-agent API

---

## Tech Stack

- **Language:** Go 1.25.6
- **LLM Providers:** AWS Bedrock (Claude), OpenAI (GPT)
- **Vector Database:** PostgreSQL with pgvector extension
- **Caching/Memory:** Redis
- **API Framework:** go-restful
- **Frontend:** Python Streamlit

## Getting Started

Each service has its own README with detailed setup instructions:

1. **[eval-agent](./eval-agent/README.md)** - Start here for LLM evaluation
2. **[search-service](./search-service/README.md)** - Required for kg-agent
3. **[kg-agent](./kg-agent/README.md)** - Depends on search-service
4. **[streamlit-ui](./streamlit-ui/README.md)** - Optional web interface

## License

MIT - See [LICENSE](./LICENSE) file for details.
