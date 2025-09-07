package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cpuchip/journal-mcp/internal/journal"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mark3labs/mcp-go/mcp"
)

// TODO: Embed frontend files when built
// For now, this is a placeholder - frontend will be embedded after npm run build
// //go:embed frontend/dist/*
// var frontendFiles embed.FS

// Placeholder for development
var frontendFiles embed.FS

// WebServer provides REST API endpoints for the web interface
type WebServer struct {
	journalService *journal.Service
	server         *http.Server
	upgrader       websocket.Upgrader
}

// NewWebServer creates a new web server instance  
func NewServer(journalService *journal.Service, port int) *WebServer {
	ws := &WebServer{
		journalService: journalService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development - in production, restrict this
				return true
			},
		},
	}

	router := mux.NewRouter()
	ws.setupRoutes(router)

	ws.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: ws.corsMiddleware(router),
	}

	return ws
}

// Start starts the web server
func (ws *WebServer) Start() error {
	log.Printf("Starting web server on %s", ws.server.Addr)
	return ws.server.ListenAndServe()
}

// Stop stops the web server
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.server.Shutdown(ctx)
}

// CORS middleware to allow cross-origin requests
func (ws *WebServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Setup all REST API routes
func (ws *WebServer) setupRoutes(router *mux.Router) {
	api := router.PathPrefix("/api").Subrouter()

	// Task endpoints
	api.HandleFunc("/tasks", ws.handleGetTasks).Methods("GET")
	api.HandleFunc("/tasks", ws.handleCreateTask).Methods("POST")
	api.HandleFunc("/tasks/{id}", ws.handleGetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", ws.handleUpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", ws.handleDeleteTask).Methods("DELETE")
	api.HandleFunc("/tasks/{id}/entries", ws.handleCreateTaskEntry).Methods("POST")
	api.HandleFunc("/tasks/{id}/status", ws.handleUpdateTaskStatus).Methods("PUT")

	// Search endpoints
	api.HandleFunc("/search", ws.handleSearch).Methods("GET")

	// Analytics endpoints
	api.HandleFunc("/analytics/overview", ws.handleAnalyticsOverview).Methods("GET")
	api.HandleFunc("/analytics/report", ws.handleAnalyticsReport).Methods("GET")

	// Export endpoints
	api.HandleFunc("/export", ws.handleExport).Methods("GET")

	// Daily/Weekly logs
	api.HandleFunc("/logs/daily/{date}", ws.handleGetDailyLog).Methods("GET")
	api.HandleFunc("/logs/weekly/{date}", ws.handleGetWeeklyLog).Methods("GET")

	// One-on-One endpoints
	api.HandleFunc("/one-on-ones", ws.handleGetOneOnOnes).Methods("GET")
	api.HandleFunc("/one-on-ones", ws.handleCreateOneOnOne).Methods("POST")

	// GitHub integration endpoints
	api.HandleFunc("/github/sync", ws.handleGitHubSync).Methods("POST")
	api.HandleFunc("/github/pull-updates", ws.handlePullIssueUpdates).Methods("POST")
	api.HandleFunc("/github/create-task", ws.handleCreateTaskFromGitHub).Methods("POST")

	// Data management endpoints
	api.HandleFunc("/backup", ws.handleCreateBackup).Methods("POST")
	api.HandleFunc("/restore", ws.handleRestoreBackup).Methods("POST")
	api.HandleFunc("/config", ws.handleGetConfig).Methods("GET")
	api.HandleFunc("/config", ws.handleUpdateConfig).Methods("PUT")

	// WebSocket endpoint for real-time updates
	api.HandleFunc("/ws", ws.handleWebSocket)

	// Serve OpenAPI documentation
	api.HandleFunc("/docs", ws.handleAPIDocs).Methods("GET")

	// Health check
	api.HandleFunc("/health", ws.handleHealth).Methods("GET")
	
	// Serve embedded frontend files
	ws.setupStaticRoutes(router)
}

// setupStaticRoutes serves the embedded Vue.js frontend
func (ws *WebServer) setupStaticRoutes(router *mux.Router) {
	// Get embedded filesystem
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Printf("Warning: Could not access embedded frontend files: %v", err)
		return
	}

	// Create file server for static assets
	fileServer := http.FileServer(http.FS(frontendFS))

	// Handle static files (CSS, JS, images)
	router.PathPrefix("/assets/").Handler(fileServer)
	
	// Handle all other routes (SPA routing)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't serve frontend for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the requested file
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		if _, err := frontendFS.Open(strings.TrimPrefix(path, "/")); err != nil {
			// File doesn't exist, serve index.html for SPA routing
			path = "/index.html"
		}

		// Set appropriate content type
		if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}

		// Serve the file
		http.ServeFile(w, r, path)
	})
}

// Task Handlers

func (ws *WebServer) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	// Convert query parameters to MCP request format
	args := make(map[string]interface{})
	if status := query.Get("status"); status != "" {
		args["status"] = status
	}
	if taskType := query.Get("type"); taskType != "" {
		args["type"] = taskType
	}
	if dateFrom := query.Get("date_from"); dateFrom != "" {
		args["date_from"] = dateFrom
	}
	if dateTo := query.Get("date_to"); dateTo != "" {
		args["date_to"] = dateTo
	}
	if limit := query.Get("limit"); limit != "" {
		args["limit"] = limit
	}
	if offset := query.Get("offset"); offset != "" {
		args["offset"] = offset
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.ListTasks(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var taskData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request := createMCPRequest(taskData)
	result, err := ws.journalService.CreateTask(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	args := map[string]interface{}{
		"task_id": taskID,
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.GetTask(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updateData["task_id"] = taskID

	// This would need a new MCP tool for updating task metadata
	// For now, we'll return a not implemented error
	http.Error(w, "Task update not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	// This would need a new MCP tool for deleting tasks
	// For now, we'll return a not implemented error
	http.Error(w, "Task deletion not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleCreateTaskEntry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var entryData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&entryData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	entryData["task_id"] = taskID

	request := createMCPRequest(entryData)
	result, err := ws.journalService.AddTaskEntry(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleUpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var statusData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&statusData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	statusData["task_id"] = taskID

	request := createMCPRequest(statusData)
	result, err := ws.journalService.UpdateTaskStatus(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

// Search Handlers

func (ws *WebServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	args := map[string]interface{}{
		"query": query.Get("q"),
	}
	if taskType := query.Get("task_type"); taskType != "" {
		args["task_type"] = taskType
	}
	if dateFrom := query.Get("date_from"); dateFrom != "" {
		args["date_from"] = dateFrom
	}
	if dateTo := query.Get("date_to"); dateTo != "" {
		args["date_to"] = dateTo
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.SearchEntries(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

// Analytics Handlers

func (ws *WebServer) handleAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	args := map[string]interface{}{
		"report_type": "overview",
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.GetAnalyticsReport(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleAnalyticsReport(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	args := map[string]interface{}{}
	if reportType := query.Get("type"); reportType != "" {
		args["report_type"] = reportType
	}
	if timePeriod := query.Get("period"); timePeriod != "" {
		args["time_period"] = timePeriod
	}
	if taskType := query.Get("task_type"); taskType != "" {
		args["task_type"] = taskType
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.GetAnalyticsReport(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

// Export Handler

func (ws *WebServer) handleExport(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	args := map[string]interface{}{
		"format": query.Get("format"),
	}
	if dateFrom := query.Get("date_from"); dateFrom != "" {
		args["date_from"] = dateFrom
	}
	if dateTo := query.Get("date_to"); dateTo != "" {
		args["date_to"] = dateTo
	}
	if taskFilter := query.Get("task_filter"); taskFilter != "" {
		args["task_filter"] = taskFilter
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.ExportData(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set appropriate content type and filename based on format
	format := query.Get("format")
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=journal-export-%s.json", time.Now().Format("2006-01-02")))
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=journal-export-%s.csv", time.Now().Format("2006-01-02")))
	case "markdown":
		w.Header().Set("Content-Type", "text/markdown")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=journal-export-%s.md", time.Now().Format("2006-01-02")))
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	ws.writeJSONResponse(w, result)
}

// Log Handlers

func (ws *WebServer) handleGetDailyLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	args := map[string]interface{}{
		"date": date,
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.GetDailyLog(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleGetWeeklyLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	args := map[string]interface{}{
		"week_start": date,
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.GetWeeklyLog(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

// One-on-One Handlers

func (ws *WebServer) handleGetOneOnOnes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	args := map[string]interface{}{}
	if limit := query.Get("limit"); limit != "" {
		args["limit"] = limit
	}

	request := createMCPRequest(args)
	result, err := ws.journalService.GetOneOnOneHistory(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleCreateOneOnOne(w http.ResponseWriter, r *http.Request) {
	var oneOnOneData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&oneOnOneData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request := createMCPRequest(oneOnOneData)
	result, err := ws.journalService.CreateOneOnOne(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	ws.writeJSONResponse(w, result)
}

// GitHub Integration Handlers

func (ws *WebServer) handleGitHubSync(w http.ResponseWriter, r *http.Request) {
	var syncData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&syncData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request := createMCPRequest(syncData)
	result, err := ws.journalService.SyncWithGitHub(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handlePullIssueUpdates(w http.ResponseWriter, r *http.Request) {
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request := createMCPRequest(updateData)
	result, err := ws.journalService.PullIssueUpdates(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws.writeJSONResponse(w, result)
}

func (ws *WebServer) handleCreateTaskFromGitHub(w http.ResponseWriter, r *http.Request) {
	var taskData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	request := createMCPRequest(taskData)
	result, err := ws.journalService.CreateTaskFromGitHubIssue(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	ws.writeJSONResponse(w, result)
}

// Data Management Handlers

func (ws *WebServer) handleCreateBackup(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with backup functionality
	http.Error(w, "Backup creation not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleRestoreBackup(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with restore functionality
	http.Error(w, "Backup restore not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with configuration management
	http.Error(w, "Configuration management not implemented yet", http.StatusNotImplemented)
}

func (ws *WebServer) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with configuration management
	http.Error(w, "Configuration management not implemented yet", http.StatusNotImplemented)
}

// WebSocket Handler for real-time updates

func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Handle WebSocket connection for real-time updates
	// This is a basic implementation - in production you'd want
	// to track connections and broadcast updates when tasks change
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Echo back for now - implement real-time update logic here
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// Utility Handlers

func (ws *WebServer) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	// Return OpenAPI/Swagger documentation
	docs := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Journal MCP REST API",
			"description": "REST API for the Journal MCP task management system",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{"url": "/api", "description": "API server"},
		},
		"paths": map[string]interface{}{
			// Simplified documentation - in production, use proper OpenAPI generator
			"/tasks": map[string]interface{}{
				"get":  map[string]interface{}{"summary": "Get all tasks"},
				"post": map[string]interface{}{"summary": "Create a new task"},
			},
			"/tasks/{id}": map[string]interface{}{"get": map[string]interface{}{"summary": "Get task by ID"}},
			"/search":     map[string]interface{}{"get": map[string]interface{}{"summary": "Search journal entries"}},
			"/analytics/overview": map[string]interface{}{"get": map[string]interface{}{"summary": "Get analytics overview"}},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "journal-mcp",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Helper functions

func (ws *WebServer) writeJSONResponse(w http.ResponseWriter, result *mcp.CallToolResult) {
	w.Header().Set("Content-Type", "application/json")
	
	if result.IsError {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": result.Content[0].(*mcp.TextContent).Text,
		})
		return
	}

	// Parse the MCP result content
	var responseData interface{}
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			if err := json.Unmarshal([]byte(textContent.Text), &responseData); err != nil {
				// If it's not valid JSON, return as plain text
				responseData = map[string]interface{}{
					"message": textContent.Text,
				}
			}
		}
	}

	json.NewEncoder(w).Encode(responseData)
}

// createMCPRequest creates an MCP request from a map of arguments
func createMCPRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}