package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Test helper to create a temporary journal service
func createTestJournalService(t *testing.T) (*JournalService, string) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create subdirectories
	os.MkdirAll(filepath.Join(tempDir, "tasks"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "daily"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "weekly"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "one-on-ones"), 0755)
	
	return &JournalService{dataDir: tempDir}, tempDir
}

// Mock MCP request helper
func createMockRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestCreateTask(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid task creation",
			args: map[string]interface{}{
				"id":       "TEST-001",
				"title":    "Test task",
				"type":     "work",
				"tags":     []string{"testing", "unit"},
				"priority": "high",
			},
			expectError: false,
		},
		{
			name: "missing id",
			args: map[string]interface{}{
				"title": "Test task",
				"type":  "work",
			},
			expectError: true,
			errorMsg:    "id is required",
		},
		{
			name: "missing title",
			args: map[string]interface{}{
				"id":   "TEST-002",
				"type": "work",
			},
			expectError: true,
			errorMsg:    "title is required",
		},
		{
			name: "missing type",
			args: map[string]interface{}{
				"id":    "TEST-003",
				"title": "Test task",
			},
			expectError: true,
			errorMsg:    "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.CreateTask(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				if len(result.Content) == 0 {
					t.Fatal("Expected error content")
				}
				content := result.Content[0].(mcp.TextContent)
				if content.Text != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				// Verify task was created
				taskID := tt.args["id"].(string)
				task, err := js.loadTask(taskID)
				if err != nil {
					t.Fatalf("Failed to load created task: %v", err)
				}
				
				if task.ID != taskID {
					t.Errorf("Expected task ID %s, got %s", taskID, task.ID)
				}
				if task.Title != tt.args["title"].(string) {
					t.Errorf("Expected title %s, got %s", tt.args["title"], task.Title)
				}
				if task.Type != tt.args["type"].(string) {
					t.Errorf("Expected type %s, got %s", tt.args["type"], task.Type)
				}
				if task.Status != "active" {
					t.Errorf("Expected status 'active', got %s", task.Status)
				}
				if len(task.Entries) != 1 {
					t.Errorf("Expected 1 entry, got %d", len(task.Entries))
				}
			}
		})
	}
}

func TestAddTaskEntry(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	// First create a task
	createReq := createMockRequest(map[string]interface{}{
		"id":    "TEST-ENTRY",
		"title": "Task for entry testing",
		"type":  "work",
	})
	_, err := js.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid entry addition",
			args: map[string]interface{}{
				"task_id": "TEST-ENTRY",
				"content": "This is a test entry",
			},
			expectError: false,
		},
		{
			name: "missing task_id",
			args: map[string]interface{}{
				"content": "This is a test entry",
			},
			expectError: true,
			errorMsg:    "task_id is required",
		},
		{
			name: "missing content",
			args: map[string]interface{}{
				"task_id": "TEST-ENTRY",
			},
			expectError: true,
			errorMsg:    "content is required",
		},
		{
			name: "non-existent task",
			args: map[string]interface{}{
				"task_id": "NON-EXISTENT",
				"content": "This should fail",
			},
			expectError: true,
			errorMsg:    "Failed to load task: open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.AddTaskEntry(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				// Verify entry was added
				task, err := js.loadTask(tt.args["task_id"].(string))
				if err != nil {
					t.Fatalf("Failed to load task: %v", err)
				}
				
				if len(task.Entries) != 2 { // Creation entry + new entry
					t.Errorf("Expected 2 entries, got %d", len(task.Entries))
				}
				
				lastEntry := task.Entries[len(task.Entries)-1]
				if lastEntry.Content != tt.args["content"].(string) {
					t.Errorf("Expected content '%s', got '%s'", tt.args["content"], lastEntry.Content)
				}
			}
		})
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	// Create a test task
	createReq := createMockRequest(map[string]interface{}{
		"id":    "TEST-STATUS",
		"title": "Task for status testing",
		"type":  "work",
	})
	_, err := js.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid status update",
			args: map[string]interface{}{
				"task_id": "TEST-STATUS",
				"status":  "completed",
				"reason":  "Testing completed",
			},
			expectError: false,
		},
		{
			name: "invalid status",
			args: map[string]interface{}{
				"task_id": "TEST-STATUS",
				"status":  "invalid",
			},
			expectError: true,
			errorMsg:    "Invalid status",
		},
		{
			name: "missing task_id",
			args: map[string]interface{}{
				"status": "completed",
			},
			expectError: true,
			errorMsg:    "task_id is required",
		},
		{
			name: "missing status",
			args: map[string]interface{}{
				"task_id": "TEST-STATUS",
			},
			expectError: true,
			errorMsg:    "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.UpdateTaskStatus(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				// Verify status was updated
				task, err := js.loadTask(tt.args["task_id"].(string))
				if err != nil {
					t.Fatalf("Failed to load task: %v", err)
				}
				
				if task.Status != tt.args["status"].(string) {
					t.Errorf("Expected status '%s', got '%s'", tt.args["status"], task.Status)
				}
			}
		})
	}
}

func TestGetDailyLog(t *testing.T) {
	js, tempDir := createTestJournalService(t)
	ctx := context.Background()

	// Create sample daily activity
	dailyActivity := DailyActivity{
		Date: "2025-01-01",
		Tasks: map[string][]Entry{
			"TEST-100": {
				{
					ID:        "entry_1",
					Timestamp: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
					Content:   "Task created",
					Type:      "creation",
				},
			},
		},
	}
	
	dailyPath := filepath.Join(tempDir, "daily", "2025-01-01.json")
	data, _ := json.MarshalIndent(dailyActivity, "", "  ")
	os.WriteFile(dailyPath, data, 0644)

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid date with existing data",
			args: map[string]interface{}{
				"date": "2025-01-01",
			},
			expectError: false,
		},
		{
			name: "valid date with no data",
			args: map[string]interface{}{
				"date": "2025-01-02",
			},
			expectError: false,
		},
		{
			name: "missing date",
			args: map[string]interface{}{},
			expectError: true,
			errorMsg:    "date is required",
		},
		{
			name: "invalid date format",
			args: map[string]interface{}{
				"date": "invalid-date",
			},
			expectError: true,
			errorMsg:    "Invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.GetDailyLog(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				content := result.Content[0].(mcp.TextContent)
				expectedDate := tt.args["date"].(string)
				if !contains(content.Text, "Daily Log: "+expectedDate) {
					t.Errorf("Expected daily log for date '%s' in content", expectedDate)
				}
			}
		})
	}
}

func TestSearchEntries(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	// Create a test task with entries
	createReq := createMockRequest(map[string]interface{}{
		"id":    "SEARCH-TEST",
		"title": "Search testing task",
		"type":  "work",
	})
	_, err := js.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	// Add an entry
	entryReq := createMockRequest(map[string]interface{}{
		"task_id": "SEARCH-TEST",
		"content": "This is searchable content with keyword important",
	})
	_, err = js.AddTaskEntry(ctx, entryReq)
	if err != nil {
		t.Fatalf("Failed to add test entry: %v", err)
	}

	tests := []struct {
		name         string
		args         map[string]interface{}
		expectError  bool
		expectMatch  bool
		errorMsg     string
	}{
		{
			name: "search with matching query",
			args: map[string]interface{}{
				"query": "important",
			},
			expectError: false,
			expectMatch: true,
		},
		{
			name: "search with non-matching query",
			args: map[string]interface{}{
				"query": "nonexistent",
			},
			expectError: false,
			expectMatch: false,
		},
		{
			name: "missing query",
			args: map[string]interface{}{},
			expectError: true,
			errorMsg:    "query is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.SearchEntries(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				content := result.Content[0].(mcp.TextContent)
				if tt.expectMatch {
					if !contains(content.Text, "SEARCH-TEST") {
						t.Error("Expected search to find the test task")
					}
				} else {
					if !contains(content.Text, "No matching entries found") {
						t.Error("Expected no matching entries message")
					}
				}
			}
		})
	}
}

func TestExportData(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	// Create a test task
	createReq := createMockRequest(map[string]interface{}{
		"id":    "EXPORT-TEST",
		"title": "Export testing task",
		"type":  "work",
	})
	_, err := js.CreateTask(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "export as JSON",
			args: map[string]interface{}{
				"format": "json",
			},
			expectError: false,
		},
		{
			name: "export as markdown",
			args: map[string]interface{}{
				"format": "markdown",
			},
			expectError: false,
		},
		{
			name: "export as CSV",
			args: map[string]interface{}{
				"format": "csv",
			},
			expectError: false,
		},
		{
			name: "invalid format",
			args: map[string]interface{}{
				"format": "pdf",
			},
			expectError: true,
			errorMsg:    "Invalid format",
		},
		{
			name: "missing format",
			args: map[string]interface{}{},
			expectError: true,
			errorMsg:    "format is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.ExportData(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				content := result.Content[0].(mcp.TextContent)
				format := tt.args["format"].(string)
				
				switch format {
				case "json":
					if !contains(content.Text, "\"tasks\"") {
						t.Error("Expected JSON export to contain tasks")
					}
				case "markdown":
					if !contains(content.Text, "# Journal Export") {
						t.Error("Expected markdown export to contain header")
					}
				case "csv":
					if !contains(content.Text, "Type,Date,Time") {
						t.Error("Expected CSV export to contain headers")
					}
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && s[:len(substr)] == substr) || 
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCreateOneOnOne(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid one-on-one creation",
			args: map[string]interface{}{
				"date":     "2025-01-01",
				"insights": []string{"Good progress", "Need more focus"},
				"todos":    []string{"Review code", "Plan sprint"},
				"feedback": []string{"Positive attitude"},
				"notes":    "Productive meeting",
			},
			expectError: false,
		},
		{
			name: "missing date",
			args: map[string]interface{}{
				"notes": "Some notes",
			},
			expectError: true,
			errorMsg:    "date is required",
		},
		{
			name: "invalid date format",
			args: map[string]interface{}{
				"date": "invalid-date",
			},
			expectError: true,
			errorMsg:    "Invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.CreateOneOnOne(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				content := result.Content[0].(mcp.TextContent)
				expectedDate := tt.args["date"].(string)
				if !contains(content.Text, expectedDate) {
					t.Errorf("Expected result to contain date '%s'", expectedDate)
				}
			}
		})
	}
}

func TestGetOneOnOneHistory(t *testing.T) {
	js, tempDir := createTestJournalService(t)
	ctx := context.Background()

	// Create a sample one-on-one file
	oneOnOne := OneOnOne{
		Date:     "2025-01-01",
		Insights: []string{"Great progress"},
		Todos:    []string{"Review code"},
		Notes:    "Productive meeting",
		Created:  time.Date(2025, 1, 1, 16, 0, 0, 0, time.UTC),
	}
	
	oneOnOnePath := filepath.Join(tempDir, "one-on-ones", "2025-01-01.json")
	data, _ := json.MarshalIndent(oneOnOne, "", "  ")
	os.WriteFile(oneOnOnePath, data, 0644)

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "get history with default limit",
			args: map[string]interface{}{},
		},
		{
			name: "get history with custom limit",
			args: map[string]interface{}{
				"limit": "5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.GetOneOnOneHistory(ctx, req)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.IsError {
				t.Fatal("Unexpected error result")
			}
			
			content := result.Content[0].(mcp.TextContent)
			if !contains(content.Text, "One-on-One History") {
				t.Error("Expected one-on-one history header")
			}
			if !contains(content.Text, "2025-01-01") {
				t.Error("Expected to find the test meeting date")
			}
		})
	}
}

func TestGetWeeklyLog(t *testing.T) {
	js, tempDir := createTestJournalService(t)
	ctx := context.Background()

	// Create a sample task and daily activity
	task := Task{
		ID:      "WEEKLY-TEST",
		Title:   "Weekly test task",
		Type:    "work",
		Status:  "active",
		Created: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		Updated: time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC),
		Entries: []Entry{
			{
				ID:        "entry_1",
				Timestamp: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
				Content:   "Task created",
				Type:      "creation",
			},
		},
	}
	
	taskPath := filepath.Join(tempDir, "tasks", "WEEKLY-TEST.json")
	taskData, _ := json.MarshalIndent(task, "", "  ")
	os.WriteFile(taskPath, taskData, 0644)

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid weekly log",
			args: map[string]interface{}{
				"week_start": "2025-01-01",
			},
			expectError: false,
		},
		{
			name: "missing week_start",
			args: map[string]interface{}{},
			expectError: true,
			errorMsg:    "week_start is required",
		},
		{
			name: "invalid date format",
			args: map[string]interface{}{
				"week_start": "invalid-date",
			},
			expectError: true,
			errorMsg:    "Invalid week_start format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.GetWeeklyLog(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, content.Text)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatal("Unexpected error result")
				}
				
				content := result.Content[0].(mcp.TextContent)
				if !contains(content.Text, "Weekly Log:") {
					t.Error("Expected weekly log header")
				}
				if !contains(content.Text, "Weekly Summary") {
					t.Error("Expected weekly summary section")
				}
			}
		})
	}
}

// TestDateFiltering tests the date filtering functionality in list_tasks
func TestDateFiltering(t *testing.T) {
	js, tempDir := createTestJournalService(t)
	ctx := context.Background()

	// Create test tasks with different dates
	tasks := []Task{
		{
			ID:      "TASK-OLD",
			Title:   "Old task",
			Type:    "work",
			Status:  "active",
			Created: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			Updated: time.Date(2025, 1, 1, 15, 0, 0, 0, time.UTC),
			Entries: []Entry{},
		},
		{
			ID:      "TASK-MIDDLE",
			Title:   "Middle task",
			Type:    "work",
			Status:  "active",
			Created: time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC),
			Updated: time.Date(2025, 1, 5, 15, 0, 0, 0, time.UTC),
			Entries: []Entry{},
		},
		{
			ID:      "TASK-NEW",
			Title:   "New task",
			Type:    "work",
			Status:  "active",
			Created: time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC),
			Updated: time.Date(2025, 1, 10, 15, 0, 0, 0, time.UTC),
			Entries: []Entry{},
		},
	}

	// Save test tasks
	for _, task := range tasks {
		taskPath := filepath.Join(tempDir, "tasks", task.ID+".json")
		taskData, _ := json.MarshalIndent(task, "", "  ")
		os.WriteFile(taskPath, taskData, 0644)
	}

	tests := []struct {
		name           string
		args           map[string]interface{}
		expectedTasks  []string // Task IDs that should be included
		expectError    bool
	}{
		{
			name: "filter from date_from only",
			args: map[string]interface{}{
				"date_from": "2025-01-05",
			},
			expectedTasks: []string{"TASK-MIDDLE", "TASK-NEW"}, // Should exclude TASK-OLD
		},
		{
			name: "filter with date_to only",
			args: map[string]interface{}{
				"date_to": "2025-01-05",
			},
			expectedTasks: []string{"TASK-OLD", "TASK-MIDDLE"}, // Should exclude TASK-NEW
		},
		{
			name: "filter with date range",
			args: map[string]interface{}{
				"date_from": "2025-01-02",
				"date_to":   "2025-01-08",
			},
			expectedTasks: []string{"TASK-MIDDLE"}, // Only middle task
		},
		{
			name: "no date filters",
			args: map[string]interface{}{},
			expectedTasks: []string{"TASK-OLD", "TASK-MIDDLE", "TASK-NEW"}, // All tasks
		},
		{
			name: "invalid date_from format (should ignore filter)",
			args: map[string]interface{}{
				"date_from": "invalid-date",
			},
			expectedTasks: []string{"TASK-OLD", "TASK-MIDDLE", "TASK-NEW"}, // All tasks (invalid date ignored)
		},
		{
			name: "invalid date_to format (should ignore filter)",
			args: map[string]interface{}{
				"date_to": "invalid-date",
			},
			expectedTasks: []string{"TASK-OLD", "TASK-MIDDLE", "TASK-NEW"}, // All tasks (invalid date ignored)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.ListTasks(ctx, req)

			if tt.expectError {
				if err != nil {
					t.Fatalf("Expected error in result, got actual error: %v", err)
				}
				if !result.IsError {
					t.Fatal("Expected error result")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.IsError {
					content := result.Content[0].(mcp.TextContent)
					t.Fatalf("Unexpected error result: %s", content.Text)
				}
				
				content := result.Content[0].(mcp.TextContent)
				resultText := content.Text
				
				// Check that expected tasks are present
				for _, expectedTask := range tt.expectedTasks {
					if !contains(resultText, expectedTask) {
						t.Errorf("Expected task '%s' to be in results, but it wasn't found", expectedTask)
					}
				}
				
				// Check that only expected tasks are present
				allTaskIDs := []string{"TASK-OLD", "TASK-MIDDLE", "TASK-NEW"}
				for _, taskID := range allTaskIDs {
					isExpected := false
					for _, expectedTask := range tt.expectedTasks {
						if taskID == expectedTask {
							isExpected = true
							break
						}
					}
					if !isExpected && contains(resultText, taskID) {
						t.Errorf("Unexpected task '%s' found in results", taskID)
					}
				}
			}
		})
	}
}
// TestPagination tests the pagination functionality in list_tasks
func TestPagination(t *testing.T) {
	js, tempDir := createTestJournalService(t)
	ctx := context.Background()

	// Create 5 test tasks
	tasks := []Task{}
	for i := 1; i <= 5; i++ {
		task := Task{
			ID:      fmt.Sprintf("TASK-%d", i),
			Title:   fmt.Sprintf("Task %d", i),
			Type:    "work",
			Status:  "active",
			Created: time.Date(2025, 1, i, 10, 0, 0, 0, time.UTC),
			Updated: time.Date(2025, 1, i, 15, 0, 0, 0, time.UTC),
			Entries: []Entry{},
		}
		tasks = append(tasks, task)
		
		// Save task
		taskPath := filepath.Join(tempDir, "tasks", task.ID+".json")
		taskData, _ := json.MarshalIndent(task, "", "  ")
		os.WriteFile(taskPath, taskData, 0644)
	}

	tests := []struct {
		name           string
		args           map[string]interface{}
		expectedCount  int
		shouldContain  []string
	}{
		{
			name: "default pagination (no limit/offset)",
			args: map[string]interface{}{},
			expectedCount: 5,
			shouldContain: []string{"TASK-1", "TASK-2", "TASK-3", "TASK-4", "TASK-5"},
		},
		{
			name: "limit 2 tasks",
			args: map[string]interface{}{
				"limit": "2",
			},
			expectedCount: 2,
			shouldContain: []string{"TASK-5", "TASK-4"}, // Most recent first
		},
		{
			name: "offset 2, limit 2",
			args: map[string]interface{}{
				"offset": "2",
				"limit":  "2",
			},
			expectedCount: 2,
			shouldContain: []string{"TASK-3", "TASK-2"},
		},
		{
			name: "offset beyond total (should return empty)",
			args: map[string]interface{}{
				"offset": "10",
			},
			expectedCount: 0,
			shouldContain: []string{},
		},
		{
			name: "large limit (should be capped at total)",
			args: map[string]interface{}{
				"limit": "250", // Should be capped at 200, but we only have 5 tasks
			},
			expectedCount: 5,
			shouldContain: []string{"TASK-1", "TASK-2", "TASK-3", "TASK-4", "TASK-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMockRequest(tt.args)
			result, err := js.ListTasks(ctx, req)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.IsError {
				content := result.Content[0].(mcp.TextContent)
				t.Fatalf("Unexpected error result: %s", content.Text)
			}
			
			content := result.Content[0].(mcp.TextContent)
			resultText := content.Text
			
			// Check pagination info in header
			if tt.expectedCount > 0 {
				if !contains(resultText, fmt.Sprintf("of %d total", 5)) {
					t.Errorf("Expected pagination header to show total of 5")
				}
			}
			
			// Check that expected tasks are present
			for _, expectedTask := range tt.shouldContain {
				if !contains(resultText, expectedTask) {
					t.Errorf("Expected task '%s' to be in results, but it wasn't found", expectedTask)
				}
			}
			
			// Count actual task occurrences 
			actualCount := 0
			for i := 1; i <= 5; i++ {
				taskID := fmt.Sprintf("TASK-%d", i)
				if contains(resultText, taskID) {
					actualCount++
				}
			}
			
			if actualCount != tt.expectedCount {
				t.Errorf("Expected %d tasks in results, got %d", tt.expectedCount, actualCount)
			}
		})
	}
}

func TestImportData(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		content      string
		format       string
		taskPrefix   string
		defaultType  string
		expectError  bool
		errorMsg     string
		expectTasks  int
		expectEntries int
	}{
		{
			name:          "plain text import",
			content:       "2024-01-01 First entry\nSecond entry without date\n2024-01-02 Third entry",
			format:        "txt",
			taskPrefix:    "TXT",
			defaultType:   "personal",
			expectError:   false,
			expectTasks:   1,
			expectEntries: 3,
		},
		{
			name:          "markdown import",
			content:       "# Task 1\nFirst entry\nSecond entry\n\n# Task 2\nAnother entry",
			format:        "markdown",
			taskPrefix:    "MD",
			defaultType:   "work",
			expectError:   false,
			expectTasks:   2,
			expectEntries: 3,
		},
		{
			name:        "csv import",
			content:     "title,date,content\nTask 1,2024-01-01,First entry\nTask 2,2024-01-02,Second entry",
			format:      "csv",
			taskPrefix:  "CSV",
			defaultType: "learning",
			expectError: false,
			expectTasks: 2,
			expectEntries: 2,
		},
		{
			name:        "json import",
			content:     `{"tasks": [{"id": "test", "title": "Test Task", "entries": [{"content": "Test entry", "timestamp": "2024-01-01T12:00:00Z"}]}]}`,
			format:      "json",
			taskPrefix:  "JSON",
			defaultType: "investigation",
			expectError: false,
			expectTasks: 1,
			expectEntries: 1,
		},
		{
			name:        "missing content",
			format:      "txt",
			expectError: true,
			errorMsg:    "content is required",
		},
		{
			name:        "missing format",
			content:     "test content",
			expectError: true,
			errorMsg:    "format is required",
		},
		{
			name:        "invalid format",
			content:     "test content",
			format:      "xml",
			expectError: true,
			errorMsg:    "format must be one of: txt, markdown, json, csv",
		},
		{
			name:        "invalid default type",
			content:     "test content",
			format:      "txt",
			defaultType: "invalid",
			expectError: true,
			errorMsg:    "default_type must be one of: work, learning, personal, investigation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{}
			
			if tt.content != "" {
				args["content"] = tt.content
			}
			
			if tt.format != "" {
				args["format"] = tt.format
			}
			
			if tt.taskPrefix != "" {
				args["task_prefix"] = tt.taskPrefix
			}
			if tt.defaultType != "" {
				args["default_type"] = tt.defaultType
			}

			result, err := js.ImportData(ctx, createMockRequest(args))
			
			if tt.expectError {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if !result.IsError {
					t.Error("Expected error result")
				}
				if len(result.Content) == 0 {
					t.Error("Expected error content")
				}
				content := result.Content[0].(mcp.TextContent)
				if content.Text != tt.errorMsg {
					t.Errorf("Expected error message '%s', got: %s", tt.errorMsg, content.Text)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result, got nil")
				return
			}

			if result.IsError || len(result.Content) == 0 {
				t.Error("Expected non-empty result content")
				return
			}

			content := result.Content[0].(mcp.TextContent)
			resultText := content.Text

			// Parse result as ImportResult
			var importResult ImportResult
			if err := json.Unmarshal([]byte(resultText), &importResult); err != nil {
				t.Errorf("Failed to parse import result: %v", err)
				return
			}

			if importResult.TasksCreated != tt.expectTasks {
				t.Errorf("Expected %d tasks created, got %d", tt.expectTasks, importResult.TasksCreated)
			}

			if importResult.EntriesAdded != tt.expectEntries {
				t.Errorf("Expected %d entries added, got %d", tt.expectEntries, importResult.EntriesAdded)
			}

			if importResult.Summary == "" {
				t.Error("Expected non-empty summary")
			}
		})
	}
}

func TestImportFormats(t *testing.T) {
	js, _ := createTestJournalService(t)
	ctx := context.Background()

	t.Run("plain text with dates", func(t *testing.T) {
		content := "2024-01-01 Started working on project\n" +
			"Made good progress today\n" +
			"2024-01-02 15:30 Completed the first milestone\n" +
			"01/03/2024 Final review"

		args := map[string]interface{}{
			"content":      content,
			"format":       "txt",
			"task_prefix":  "TXT",
			"default_type": "work",
		}

		result, err := js.ImportData(ctx, createMockRequest(args))
		if err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		var importResult ImportResult
		resultContent := result.Content[0].(mcp.TextContent)
		if err := json.Unmarshal([]byte(resultContent.Text), &importResult); err != nil {
			t.Fatalf("Failed to parse result: %v", err)
		}

		if importResult.TasksCreated != 1 {
			t.Errorf("Expected 1 task, got %d", importResult.TasksCreated)
		}

		if importResult.EntriesAdded != 4 {
			t.Errorf("Expected 4 entries, got %d", importResult.EntriesAdded)
		}
	})

	t.Run("markdown with multiple tasks", func(t *testing.T) {
		content := "# Project Alpha\n" +
			"Initial planning phase\n" +
			"Requirements gathering\n\n" +
			"# Project Beta\n" +
			"Started development\n" +
			"First prototype complete\n\n" +
			"# Project Gamma\n" +
			"Research phase"

		args := map[string]interface{}{
			"content":      content,
			"format":       "markdown",
			"task_prefix":  "MD",
			"default_type": "work",
		}

		result, err := js.ImportData(ctx, createMockRequest(args))
		if err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		var importResult ImportResult
		resultContent := result.Content[0].(mcp.TextContent)
		if err := json.Unmarshal([]byte(resultContent.Text), &importResult); err != nil {
			t.Fatalf("Failed to parse result: %v", err)
		}

		if importResult.TasksCreated != 3 {
			t.Errorf("Expected 3 tasks, got %d", importResult.TasksCreated)
		}

		if importResult.EntriesAdded != 5 {
			t.Errorf("Expected 5 entries, got %d", importResult.EntriesAdded)
		}
	})

	t.Run("csv with various columns", func(t *testing.T) {
		content := "title,date,content,priority\n" +
			"\"Task 1\",2024-01-01,\"First task entry\",high\n" +
			"\"Task 2\",2024-01-02,\"Second task entry\",medium\n" +
			"\"Task 1\",2024-01-03,\"Follow-up entry\",high"

		args := map[string]interface{}{
			"content":      content,
			"format":       "csv",
			"task_prefix":  "CSV",
			"default_type": "work",
		}

		result, err := js.ImportData(ctx, createMockRequest(args))
		if err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		var importResult ImportResult
		resultContent := result.Content[0].(mcp.TextContent)
		if err := json.Unmarshal([]byte(resultContent.Text), &importResult); err != nil {
			t.Fatalf("Failed to parse result: %v", err)
		}

		if importResult.TasksCreated != 2 {
			t.Errorf("Expected 2 tasks, got %d", importResult.TasksCreated)
		}

		if importResult.EntriesAdded != 3 {
			t.Errorf("Expected 3 entries, got %d", importResult.EntriesAdded)
		}
	})
}