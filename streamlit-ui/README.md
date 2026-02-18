# KG Agent Streamlit UI

Interactive chat interface for the KG Agent documentation assistant.

## Quick Start

```bash
# Install uv if you haven't already
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install dependencies
uv sync

# Run the app
uv run streamlit run app.py
```

The UI will open at `http://localhost:8501`

## Features

- 🤖 Chat interface with streaming responses
- 🟢 Haiku badge for fast model responses
- 🔵 Sonnet badge for smart model responses
- 💬 Session management and conversation history
- 📊 Token usage and cost estimation
- ⚙️ Adjustable parameters (temperature, max tokens)
- 📋 Response metadata display
- ✅ Health check indicator

## Configuration

The app connects to the KG Agent API at `http://localhost:8081/api/v1` by default.

Make sure your Go services are running:

```bash
# Terminal 1: Start search service
cd ../kg-agent
go run cmd/search/main.go

# Terminal 2: Start agent service
go run cmd/agent/main.go

# Terminal 3: Start Streamlit UI
cd ../streamlit-ui
uv run streamlit run app.py
```

## Usage

1. Type your question in the chat input
2. Watch the model badge to see which model is used:
   - 🟢 **Haiku** (Fast): Simple queries, greetings, follow-ups
   - 🔵 **Sonnet** (Smart): Complex queries, documentation search
3. View response details by expanding the "Details" section
4. Continue the conversation - session is maintained automatically
5. Use "New Session" to start a fresh conversation

## Keyboard Shortcuts

- `Enter` - Send message
- `Ctrl/Cmd + K` - Clear chat input
- `Ctrl/Cmd + R` - Refresh page

## Troubleshooting

### Agent Offline
Check the sidebar for connection status. If offline:
1. Ensure Go services are running
2. Check ports 8081 (agent) and 8082 (search)
3. Verify `.env` configuration

### Streaming Issues
If streaming doesn't work:
1. Disable "Enable Streaming" in sidebar
2. Check browser console for errors
3. Try with different browser

### Connection Timeout
Increase timeout in Settings or check:
1. Agent service is responsive
2. Database is running
3. Redis is running
