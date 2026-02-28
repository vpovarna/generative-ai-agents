.PHONY: help setup install-tools test lint clean week1 week2 week3 week4 week5

# Default target
help:
	@echo "Generative AI Agents - Production-Ready AI Agents"
	@echo ""
	@echo "Available commands:"
	@echo "  make setup         - Initial setup of all directories"
	@echo "  make install-tools - Install all required Go tools"
	@echo "  make test          - Run all tests"
	@echo "  make lint          - Run linter on all code"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Week-specific commands:"
	@echo "  make week1         - Setup Week 1 workspace"
	@echo "  make week2         - Setup Week 2 workspace"
	@echo "  make week3         - Start Week 3-4 services"
	@echo "  make week5         - Setup Week 5 AI agents"
	@echo ""
	@echo "Progress tracking:"
	@echo "  make progress      - Show learning progress"

# Initial setup
setup:
	@echo "Setting up project structure..."
	mkdir -p 01-go-fundamentals/day{1..7}
	mkdir -p 02-concurrency/{pipeline,workerpool,pubsub}
	mkdir -p 03-microservices/{order-service,search-service,worker-service}
	mkdir -p 04-ai-agents/{cmd,internal}
	@echo "✓ Project structure created"

# Install required tools
install-tools:
	@echo "Installing Go tools..."
	go install golang.org/x/tools/gopls@latest
	go install github.com/go-delve/delve/cmd/dlv@latest
	go install github.com/air-verse/air@latest
	go install go.uber.org/mock/mockgen@v0.6.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "✓ Go tools installed"
	@echo ""
	@echo "Checking other dependencies..."
	@command -v docker >/dev/null 2>&1 || echo "⚠ Docker not found. Install: brew install docker"
	@command -v docker-compose >/dev/null 2>&1 || echo "⚠ Docker Compose not found. Install: brew install docker-compose"
	@command -v protoc >/dev/null 2>&1 || echo "⚠ Protobuf compiler not found. Install: brew install protobuf"
	@command -v aws >/dev/null 2>&1 || echo "⚠ AWS CLI not found. Install: brew install awscli"
	@command -v golangci-lint >/dev/null 2>&1 || echo "⚠ golangci-lint not found. Install: brew install golangci-lint"
	@echo "✓ Dependency check complete"

# Week 1: Go Fundamentals
week1:
	@echo "Setting up Week 1: Go Fundamentals"
	@if [ ! -d "01-go-fundamentals" ]; then \
		mkdir -p 01-go-fundamentals/day{1..7}; \
	fi
	@cd 01-go-fundamentals && \
	if [ ! -f "go.mod" ]; then \
		go mod init github.com/povarna/generative-ai-agents/fundamentals; \
	fi
	@echo "✓ Week 1 ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Open 01-WEEK1-GO-FUNDAMENTALS.md"
	@echo "  2. cd 01-go-fundamentals"
	@echo "  3. Start with Day 1 problems"

# Week 2: Concurrency
week2:
	@echo "Setting up Week 2: Concurrency"
	@if [ ! -d "02-concurrency" ]; then \
		mkdir -p 02-concurrency/{pipeline,workerpool,pubsub}; \
	fi
	@cd 02-concurrency && \
	if [ ! -f "go.mod" ]; then \
		go mod init github.com/povarna/generative-ai-agents/concurrency; \
	fi
	@echo "✓ Week 2 ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Open 02-WEEK2-CONCURRENCY.md"
	@echo "  2. cd 02-concurrency"
	@echo "  3. Implement the 3 projects"

# Week 3-4: Microservices
week3:
	@echo "Setting up Week 3-4: Microservices"
	@if [ ! -d "03-microservices" ]; then \
		echo "Creating microservices structure..."; \
		mkdir -p 03-microservices/{order-service,search-service,worker-service}; \
	fi
	@echo "Starting infrastructure..."
	@cd 03-microservices && docker-compose up -d postgres redis elasticsearch 2>/dev/null || \
		echo "⚠ Docker Compose not configured yet. Follow 03-WEEK3-4-MICROSERVICES.md to create docker-compose.yml"
	@echo "✓ Week 3-4 setup initiated!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Open 03-WEEK3-4-MICROSERVICES.md"
	@echo "  2. cd 03-microservices"
	@echo "  3. Follow the guide to build services"

# Week 5: AI Agents
week5:
	@echo "Setting up Week 5: AI Agents"
	@if [ ! -d "04-ai-agents" ]; then \
		mkdir -p 04-ai-agents/{cmd/agent,internal/{llm/bedrock,mcp,agent,tools}}; \
	fi
	@cd 04-ai-agents && \
	if [ ! -f "go.mod" ]; then \
		go mod init github.com/povarna/generative-ai-agents/ai-agents; \
	fi
	@echo "Checking AWS configuration..."
	@aws sts get-caller-identity >/dev/null 2>&1 && \
		echo "✓ AWS credentials configured" || \
		echo "⚠ AWS credentials not configured. Run: aws configure"
	@echo "✓ Week 5 ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Open 04-WEEK5-AI-AGENTS.md"
	@echo "  2. Configure AWS credentials: aws configure"
	@echo "  3. Enable Bedrock models in AWS Console"
	@echo "  4. cd 04-ai-agents"
	@echo "  5. Build your first agent"

# Run all tests
test:
	@echo "Running tests across all modules..."
	@for dir in 01-go-fundamentals 02-concurrency 04-ai-agents; do \
		if [ -d $$dir ] && [ -f $$dir/go.mod ]; then \
			echo ""; \
			echo "Testing $$dir..."; \
			cd $$dir && go test -v -race ./... || true; \
			cd ..; \
		fi \
	done
	@if [ -d "03-microservices" ]; then \
		echo ""; \
		echo "Testing microservices..."; \
		cd 03-microservices && \
		for service in order-service search-service worker-service; do \
			if [ -d $$service ] && [ -f $$service/go.mod ]; then \
				echo "Testing $$service..."; \
				cd $$service && go test -v ./... || true; \
				cd ..; \
			fi \
		done; \
	fi
	@echo ""
	@echo "✓ All tests complete"

# Lint all code
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "golangci-lint not found. Install: brew install golangci-lint" && exit 1)
	@for dir in 01-go-fundamentals 02-concurrency 04-ai-agents; do \
		if [ -d $$dir ] && [ -f $$dir/go.mod ]; then \
			echo "Linting $$dir..."; \
			cd $$dir && golangci-lint run ./... || true; \
			cd ..; \
		fi \
	done
	@echo "✓ Linting complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@find . -type d -name "vendor" -exec rm -rf {} + 2>/dev/null || true
	@if [ -d "03-microservices" ]; then \
		cd 03-microservices && docker-compose down -v 2>/dev/null || true; \
	fi
	@echo "✓ Clean complete"

# Show progress
progress:
	@echo "Learning Progress Report"
	@echo "========================"
	@echo ""
	@echo "Week 1: Go Fundamentals"
	@if [ -d "01-go-fundamentals" ]; then \
		echo "  ✓ Directory exists"; \
		for day in 1 2 3 4 5 6 7; do \
			if [ -d "01-go-fundamentals/day$$day" ]; then \
				count=$$(find 01-go-fundamentals/day$$day -name "*.go" 2>/dev/null | wc -l); \
				if [ $$count -gt 0 ]; then \
					echo "  ✓ Day $$day: $$count files"; \
				else \
					echo "  ☐ Day $$day: Not started"; \
				fi \
			fi \
		done; \
	else \
		echo "  ☐ Not started"; \
	fi
	@echo ""
	@echo "Week 2: Concurrency"
	@if [ -d "02-concurrency" ]; then \
		echo "  ✓ Directory exists"; \
		for proj in pipeline workerpool pubsub; do \
			if [ -f "02-concurrency/$$proj/$$proj.go" ]; then \
				echo "  ✓ $$proj implemented"; \
			else \
				echo "  ☐ $$proj not started"; \
			fi \
		done; \
	else \
		echo "  ☐ Not started"; \
	fi
	@echo ""
	@echo "Week 3-4: Microservices"
	@if [ -d "03-microservices" ]; then \
		echo "  ✓ Directory exists"; \
		for service in order-service search-service worker-service; do \
			if [ -f "03-microservices/$$service/go.mod" ]; then \
				echo "  ✓ $$service started"; \
			else \
				echo "  ☐ $$service not started"; \
			fi \
		done; \
	else \
		echo "  ☐ Not started"; \
	fi
	@echo ""
	@echo "Week 5: AI Agents"
	@if [ -d "04-ai-agents" ]; then \
		echo "  ✓ Directory exists"; \
		if [ -f "04-ai-agents/internal/llm/bedrock/client.go" ]; then \
			echo "  ✓ Bedrock client implemented"; \
		else \
			echo "  ☐ Bedrock client not started"; \
		fi; \
		if [ -f "04-ai-agents/internal/agent/react_agent.go" ]; then \
			echo "  ✓ ReAct agent implemented"; \
		else \
			echo "  ☐ ReAct agent not started"; \
		fi; \
	else \
		echo "  ☐ Not started"; \
	fi
	@echo ""
	@echo "Overall Progress:"
	@total=0; completed=0; \
	[ -d "01-go-fundamentals" ] && total=$$((total+7)); \
	[ -d "02-concurrency" ] && total=$$((total+3)); \
	[ -d "03-microservices" ] && total=$$((total+3)); \
	[ -d "04-ai-agents" ] && total=$$((total+2)); \
	for day in 1 2 3 4 5 6 7; do \
		count=$$(find 01-go-fundamentals/day$$day -name "*.go" 2>/dev/null | wc -l); \
		[ $$count -gt 0 ] && completed=$$((completed+1)); \
	done; \
	for proj in pipeline workerpool pubsub; do \
		[ -f "02-concurrency/$$proj/$$proj.go" ] && completed=$$((completed+1)); \
	done; \
	for service in order-service search-service worker-service; do \
		[ -f "03-microservices/$$service/go.mod" ] && completed=$$((completed+1)); \
	done; \
	[ -f "04-ai-agents/internal/llm/bedrock/client.go" ] && completed=$$((completed+1)); \
	[ -f "04-ai-agents/internal/agent/react_agent.go" ] && completed=$$((completed+1)); \
	if [ $$total -gt 0 ]; then \
		percent=$$((completed * 100 / total)); \
		echo "  $$completed / $$total tasks completed ($$percent%)"; \
	fi
	@echo ""

# Quick reference
ref:
	@echo "Quick Reference"
	@echo "==============="
	@echo ""
	@echo "Common Go commands:"
	@echo "  go run main.go        - Run program"
	@echo "  go test ./...         - Run all tests"
	@echo "  go test -v ./...      - Verbose tests"
	@echo "  go test -race ./...   - Race detector"
	@echo "  go test -cover ./...  - Coverage"
	@echo "  go build              - Build binary"
	@echo "  go mod tidy           - Clean dependencies"
	@echo "  go fmt ./...          - Format code"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-compose up -d        - Start services"
	@echo "  docker-compose logs -f      - View logs"
	@echo "  docker-compose down         - Stop services"
	@echo "  docker-compose ps           - List services"
	@echo ""
	@echo "Debugging:"
	@echo "  dlv debug             - Start debugger"
	@echo "  dlv test              - Debug tests"

# Create initial .gitignore
gitignore:
	@echo "Creating .gitignore..."
	@cat > .gitignore <<EOF
	# Binaries
	*.exe
	*.exe~
	*.dll
	*.so
	*.dylib
	*.test
	*.out
	
	# Go workspace file
	go.work
	
	# Dependency directories
	vendor/
	
	# IDEs
	.vscode/
	.idea/
	*.swp
	*.swo
	*~
	.DS_Store
	
	# Environment variables
	.env
	.env.local
	
	# Test coverage
	coverage.txt
	coverage.html
	*.cover
	
	# Build output
	bin/
	dist/
	
	# Docker
	docker-compose.override.yml
	
	# Logs
	*.log
	
	# OS files
	Thumbs.db
	
	# Temporary files
	tmp/
	temp/
	EOF
	@echo "✓ .gitignore created"

# Bootstrap everything
bootstrap: setup install-tools gitignore
	@echo ""
	@echo "✓ Bootstrap complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Read README.md for overview"
	@echo "  2. Open 00-SETUP.md for detailed setup"
	@echo "  3. Run 'make week1' to start Week 1"
	@echo ""
	@echo "Happy learning! 🚀"
