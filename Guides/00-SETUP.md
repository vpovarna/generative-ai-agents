# Go Development Setup in Cursor

## Prerequisites

### 1. Install Go
```bash
# macOS (using Homebrew)
brew install go

# Verify installation
go version  # Should show 1.24+ (or latest 1.23.x)
```

### 2. Go Environment Setup
```bash
# Add to ~/.zshrc
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Create workspace
mkdir -p $HOME/go/{bin,src,pkg}
```

### 3. Essential Go Tools

```bash
# Language server (gopls) - for IntelliSense
go install golang.org/x/tools/gopls@latest

# Linter
brew install golangci-lint

# Test coverage
go install github.com/axw/gocov/gocov@latest

# Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Air - live reload for development
go install github.com/air-verse/air@latest

# Mock generation (from your stack)
go install go.uber.org/mock/mockgen@v0.6.0

# Protobuf tools
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Database migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Cursor Extensions/Settings

### Recommended Extensions
1. **Go** (Go Team at Google) - should auto-install when you open .go files
2. **Protocol Buffers** - for .proto files
3. **Docker** - for containerization
4. **REST Client** - for API testing

### Cursor Settings for Go

Create `.vscode/settings.json` in your workspace:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.testTimeout": "60s",
  "go.coverOnSave": false,
  "go.testFlags": ["-v", "-race"],
  "go.buildTags": "",
  "go.toolsManagement.autoUpdate": true,
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "go.testEnvVars": {},
  "go.coverMode": "atomic"
}
```

### Project Structure
```
Generative_AI_With_GO/
├── 01-go-fundamentals/      # Week 1: AoC problems
├── 02-concurrency/           # Week 2: Goroutines, channels
├── 03-microservices/         # Week 3-4: REST/gRPC
│   ├── api-service/
│   ├── worker-service/
│   └── shared/
└── 04-ai-agents/             # Week 5: MCP + Bedrock
```

## Quick Start Test

Create a test file to verify everything works:

```bash
mkdir -p 01-go-fundamentals
cd 01-go-fundamentals
go mod init github.com/povarna/generative-ai-with-go/fundamentals
```

Create `main.go`:
```go
package main

import "fmt"

func main() {
    fmt.Println("Go development environment ready! 🚀")
}
```

Run:
```bash
go run main.go
```

## Debugging Setup

Cursor uses VS Code's debugging protocol. Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${fileDirname}"
    },
    {
      "name": "Launch Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}"
    }
  ]
}
```