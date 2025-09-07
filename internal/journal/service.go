package journal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// NewService creates a new journal service instance
func NewService() *Service {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".journal-mcp")

	// Ensure directories exist
	os.MkdirAll(filepath.Join(dataDir, "tasks"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "daily"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "weekly"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "one-on-ones"), 0755)

	return &Service{
		dataDir: dataDir,
	}
}

// PLACEHOLDER - Need to extract and properly format all methods from journal.go, github.go, and config.go
// This will be a placeholder for now while we test the structure

// Stub implementations to test the structure - need to implement properly
func (js *Service) CreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("id is required"), nil
	}

	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError("title is required"), nil
	}

	taskType, err := request.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError("type is required"), nil
	}

	// Parse tags if provided
	var tags []string
	if tagsSlice := request.GetStringSlice("tags", nil); tagsSlice != nil {
		tags = tagsSlice
	}

	task := Task{
		ID:      id,
		Title:   title,
		Type:    taskType,
		Tags:    tags,
		Status:  "active",
		Created: time.Now(),
		Updated: time.Now(),
		Entries: []Entry{},
	}

	// Handle optional fields
	if priority := request.GetString("priority", ""); priority != "" {
		task.Priority = priority
	}

	if issueURL := request.GetString("issue_url", ""); issueURL != "" {
		task.IssueURL = issueURL
		// Extract issue ID from URL for easier referencing
		if strings.Contains(issueURL, "github.com") {
			parts := strings.Split(issueURL, "/")
			if len(parts) >= 2 {
				task.IssueID = parts[len(parts)-1]
			}
		} else if strings.Contains(issueURL, "jira") || strings.Contains(issueURL, "atlassian") {
			// Extract Jira ticket ID
			parts := strings.Split(issueURL, "/")
			for _, part := range parts {
				if strings.Contains(part, "-") && len(part) > 3 {
					task.IssueID = part
					break
				}
			}
		}
	}

	// Add creation entry
	task.Entries = append(task.Entries, Entry{
		ID:        generateEntryID(),
		Timestamp: time.Now(),
		Content:   fmt.Sprintf("Task created: %s", title),
		Type:      "creation",
	})

	// Save task
	if err := js.saveTask(&task); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to save task: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Created task %s: %s", id, title),
			},
		},
	}, nil
}

func (js *Service) AddTaskEntry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("AddTaskEntry stub - needs implementation"), nil
}

func (js *Service) GetTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetTask stub - needs implementation"), nil
}

func (js *Service) ListTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ListTasks stub - needs implementation"), nil
}

func (js *Service) UpdateTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("UpdateTaskStatus stub - needs implementation"), nil
}

func (js *Service) GetDailyLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetDailyLog stub - needs implementation"), nil
}

func (js *Service) GetWeeklyLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetWeeklyLog stub - needs implementation"), nil
}

func (js *Service) CreateOneOnOne(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("CreateOneOnOne stub - needs implementation"), nil
}

func (js *Service) GetOneOnOneHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetOneOnOneHistory stub - needs implementation"), nil
}

func (js *Service) SearchEntries(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("SearchEntries stub - needs implementation"), nil
}

func (js *Service) ExportData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ExportData stub - needs implementation"), nil
}

func (js *Service) ImportData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ImportData stub - needs implementation"), nil
}

func (js *Service) GetTaskRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetTaskRecommendations stub - needs implementation"), nil
}

func (js *Service) GetAnalyticsReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetAnalyticsReport stub - needs implementation"), nil
}

func (js *Service) SyncWithGitHub(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("SyncWithGitHub stub - needs implementation"), nil
}

func (js *Service) PullIssueUpdates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("PullIssueUpdates stub - needs implementation"), nil
}

func (js *Service) CreateTaskFromGitHubIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("CreateTaskFromGitHubIssue stub - needs implementation"), nil
}

func (js *Service) CreateDataBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("CreateDataBackup stub - needs implementation"), nil
}

func (js *Service) RestoreDataBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("RestoreDataBackup stub - needs implementation"), nil
}

func (js *Service) GetConfiguration(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("GetConfiguration stub - needs implementation"), nil
}

func (js *Service) UpdateConfiguration(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("UpdateConfiguration stub - needs implementation"), nil
}

func (js *Service) MigrateData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("MigrateData stub - needs implementation"), nil
}

// Helper functions
func generateEntryID() string {
	return fmt.Sprintf("entry_%d", time.Now().UnixNano())
}

func (js *Service) saveTask(task *Task) error {
	filePath := filepath.Join(js.dataDir, "tasks", task.ID+".json")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (js *Service) loadTask(taskID string) (*Task, error) {
	filePath := filepath.Join(js.dataDir, "tasks", taskID+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}