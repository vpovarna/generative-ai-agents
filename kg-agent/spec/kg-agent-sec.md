## Scope 

Create a KG Agent. It should use Bedrock / Claude for reasoning.
The communication method should be API base. This should be an application which can start and accept messages through an API POST request. 
Is should have tools, connect to DB or to another API for fetching similarities. Performing hybrid search. 
Maybe should have other tools.
I would also like to add guardrails and query rewrite as tasks in the flow. 

## Phases
Phase 1: Foundation - AWS/Claude Connection & Basic LLM  
  - Set up AWS Bedrock client in Go
  - Implement basic Claude API integration
  - Create simple prompt/response flow
  - Test basic reasoning capabilities
  - Deliverable: A CLI or simple function that can send a prompt to Claude and get a response

Phase 2: API Layer  
  - Build HTTP server with POST endpoint for queries
  - Request/response models for the agent
  - Basic error handling and logging
  - Deliverable: REST API that accepts documentation questions and returns Claude responses

Phase 3: Query Write logic
  - Add query rewriting for better retrieval

Phase 4: Knowledge Base & Vector Search  
  - Connect to vector database (e.g., PostgreSQL with pgvector, or Amazon OpenSearch)
  - Implement document ingestion pipeline
  - Build similarity search tool
  - Implement hybrid search (keyword + semantic)
  - Deliverable: Agent can retrieve relevant documentation chunks before answering

Phase 5: Agent Tools & Reasoning Loop  
  - Implement tool-calling framework
  - Create tools: document_search, get_related_docs, etc.
  - Add ReAct or similar reasoning pattern
  - Deliverable: Agent can decide when to search and use retrieved context

Phase 6: Advanced Features  
  - Implement guardrails (input/output validation, safety checks)
  - Add conversation memory/history
  - Deliverable: Production-ready documentation agent