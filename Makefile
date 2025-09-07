.PHONY: build test clean dev frontend frontend-dev install-deps help

# Default target
all: build

# Build the complete application
build: frontend
	@echo "Building Journal MCP..."
	go build -o journal-mcp ./cmd/journal-mcp
	@echo "✅ Build complete! Binary: journal-mcp"

# Build frontend only
frontend:
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "✅ Frontend built successfully"

# Development mode with hot reload
dev:
	@echo "Starting development mode..."
	@echo "Frontend: http://localhost:5173"
	@echo "Backend API: http://localhost:8080"
	@$(MAKE) -j2 frontend-dev backend-dev

# Start frontend development server
frontend-dev:
	cd frontend && npm run dev

# Start backend in web mode
backend-dev:
	go run ./cmd/journal-mcp --web

# Run tests
test:
	@echo "Running Go tests..."
	go test -v ./...
	@echo "✅ Tests completed"

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f journal-mcp journal-mcp.exe
	rm -rf frontend/dist
	rm -rf frontend/node_modules/.cache
	@echo "✅ Clean completed"

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "✅ Dependencies installed"

# Install frontend dependencies only
install-frontend-deps:
	cd frontend && npm install

# Format code
fmt:
	go fmt ./...
	cd frontend && npm run format 2>/dev/null || echo "No frontend format script"

# Lint code
lint:
	go vet ./...
	cd frontend && npm run lint 2>/dev/null || echo "No frontend lint script"

# Create a production build with embedded frontend
build-prod: frontend
	@echo "Building production binary with embedded frontend..."
	go build -ldflags="-s -w" -o journal-mcp ./cmd/journal-mcp
	@echo "✅ Production build complete!"

# Quick development setup
setup: install-deps
	@echo "🚀 Development setup complete!"
	@echo ""
	@echo "To start developing:"
	@echo "  make dev       # Start both frontend and backend in dev mode"
	@echo "  make build     # Build the complete application"
	@echo "  make test      # Run tests"

# Show available targets
help:
	@echo "Journal MCP - Available Make Targets:"
	@echo ""
	@echo "  build          Build the complete application (frontend + backend)"
	@echo "  frontend       Build frontend only"
	@echo "  dev            Start development mode (both frontend and backend)"
	@echo "  test           Run Go tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  clean          Clean build artifacts"
	@echo "  install-deps   Install all dependencies (Go + npm)"
	@echo "  setup          Complete development setup"
	@echo "  build-prod     Create optimized production build"
	@echo "  fmt            Format code"
	@echo "  lint           Lint code"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Development workflow:"
	@echo "  1. make setup     # First time setup"
	@echo "  2. make dev       # Start development"
	@echo "  3. make build     # Create production build"