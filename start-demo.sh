#!/bin/bash
# Start the Journal MCP web server for demonstration

echo "Starting Journal MCP Web Server Demo..."
echo "======================================="

# Check if journal-mcp exists
if [ ! -f "./journal-mcp" ]; then
    echo "Building journal-mcp..."
    go build -o journal-mcp
fi

echo ""
echo "🚀 Starting Journal MCP in web mode..."
echo "   API Server: http://localhost:8080/api"
echo "   API Docs:   http://localhost:8080/api/docs"
echo "   Demo Page:  file://$(pwd)/demo.html"
echo ""
echo "📋 Available endpoints:"
echo "   GET  /api/health          - Health check"
echo "   GET  /api/docs            - API documentation"
echo "   GET  /api/tasks           - List tasks"
echo "   POST /api/tasks           - Create task"
echo "   GET  /api/search          - Search entries"
echo "   GET  /api/analytics/*     - Analytics"
echo "   POST /api/github/sync     - GitHub sync"
echo "   POST /api/backup          - Create backup"
echo "   WS   /api/ws              - WebSocket"
echo ""
echo "🌐 Open demo.html in your browser to interact with the API"
echo ""
echo "Press Ctrl+C to stop the server"
echo "======================================="

# Start the web server
./journal-mcp --web