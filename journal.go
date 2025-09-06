package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type JournalService struct {
	dataDir string
}

type Task struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Type     string    `json:"type"` // work, learning, personal, investigation
	Tags     []string  `json:"tags"`
	Status   string    `json:"status"` // active, completed, paused, blocked
	Priority string    `json:"priority,omitempty"`
	IssueURL string    `json:"issue_url,omitempty"`
	IssueID  string    `json:"issue_id,omitempty"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	Entries  []Entry   `json:"entries"`
}

type Entry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
	Type      string    `json:"type,omitempty"` // log, status_change, completion, etc.
}

type OneOnOne struct {
	Date     string    `json:"date"`
	Insights []string  `json:"insights,omitempty"`
	Todos    []string  `json:"todos,omitempty"`
	Feedback []string  `json:"feedback,omitempty"`
	Notes    string    `json:"notes,omitempty"`
	Created  time.Time `json:"created"`
}

type DailyActivity struct {
	Date  string             `json:"date"`
	Tasks map[string][]Entry `json:"tasks"` // task_id -> entries for that day
}

func NewJournalService() *JournalService {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".journal-mcp")

	// Ensure directories exist
	os.MkdirAll(filepath.Join(dataDir, "tasks"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "daily"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "weekly"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "one-on-ones"), 0755)

	return &JournalService{
		dataDir: dataDir,
	}
}

func (js *JournalService) CreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func (js *JournalService) AddTaskEntry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, err := request.RequireString("task_id")
	if err != nil {
		return mcp.NewToolResultError("task_id is required"), nil
	}

	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	// Load existing task
	task, err := js.loadTask(taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load task: %v", err)), nil
	}

	// Parse timestamp or use current time
	timestamp := time.Now()
	if timestampStr := request.GetString("timestamp", ""); timestampStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			timestamp = parsedTime
		}
	}

	// Add new entry
	entry := Entry{
		ID:        generateEntryID(),
		Timestamp: timestamp,
		Content:   content,
		Type:      "log",
	}

	task.Entries = append(task.Entries, entry)
	task.Updated = time.Now()

	// Save updated task
	if err := js.saveTask(task); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save task: %v", err)), nil
	}

	// Update daily log
	js.updateDailyLog(taskID, entry)

	return mcp.NewToolResultText(fmt.Sprintf("Added entry to task %s at %s", taskID, timestamp.Format("15:04"))), nil
}

func (js *JournalService) UpdateTaskEntry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, err := request.RequireString("task_id")
	if err != nil {
		return mcp.NewToolResultError("task_id is required"), nil
	}

	entryID, err := request.RequireString("entry_id")
	if err != nil {
		return mcp.NewToolResultError("entry_id is required"), nil
	}

	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	// Load task
	task, err := js.loadTask(taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load task: %v", err)), nil
	}

	// Find and update entry
	found := false
	for i, entry := range task.Entries {
		if entry.ID == entryID {
			task.Entries[i].Content = content
			task.Updated = time.Now()
			found = true
			break
		}
	}

	if !found {
		return mcp.NewToolResultError("Entry not found"), nil
	}

	// Save updated task
	if err := js.saveTask(task); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save task: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Updated entry in task %s", taskID)), nil
}

func (js *JournalService) GetTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, err := request.RequireString("task_id")
	if err != nil {
		return mcp.NewToolResultError("task_id is required"), nil
	}

	task, err := js.loadTask(taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load task: %v", err)), nil
	}

	// Format task as markdown for easy reading
	markdown := js.formatTaskAsMarkdown(task)

	return mcp.NewToolResultText(markdown), nil
}

func (js *JournalService) ListTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tasks, err := js.loadAllTasks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
	}

	// Apply filters (use request.GetArguments() to get raw map for filtering)
	filtered := js.filterTasks(tasks, request.GetArguments())

	// Sort by updated time (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Updated.After(filtered[j].Updated)
	})

	// Format as list
	var result strings.Builder
	result.WriteString("# Task List\n\n")

	for _, task := range filtered {
		result.WriteString(fmt.Sprintf("## %s: %s\n", task.ID, task.Title))
		result.WriteString(fmt.Sprintf("**Type:** %s | **Status:** %s", task.Type, task.Status))
		if task.Priority != "" {
			result.WriteString(fmt.Sprintf(" | **Priority:** %s", task.Priority))
		}
		result.WriteString("\n")
		if len(task.Tags) > 0 {
			result.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(task.Tags, ", ")))
		}
		if task.IssueURL != "" {
			result.WriteString(fmt.Sprintf("**Issue:** [%s](%s)\n", task.IssueID, task.IssueURL))
		}
		result.WriteString(fmt.Sprintf("**Updated:** %s\n\n", task.Updated.Format("2006-01-02 15:04")))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (js *JournalService) UpdateTaskStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, err := request.RequireString("task_id")
	if err != nil {
		return mcp.NewToolResultError("task_id is required"), nil
	}

	status, err := request.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError("status is required"), nil
	}

	// Validate status
	validStatuses := []string{"active", "completed", "paused", "blocked"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return mcp.NewToolResultError("Invalid status. Must be: active, completed, paused, blocked"), nil
	}

	// Load task
	task, err := js.loadTask(taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load task: %v", err)), nil
	}

	oldStatus := task.Status
	task.Status = status
	task.Updated = time.Now()

	// Add status change entry
	content := fmt.Sprintf("Status changed from %s to %s", oldStatus, status)
	if reason := request.GetString("reason", ""); reason != "" {
		content += fmt.Sprintf(": %s", reason)
	}

	entry := Entry{
		ID:        generateEntryID(),
		Timestamp: time.Now(),
		Content:   content,
		Type:      "status_change",
	}

	task.Entries = append(task.Entries, entry)

	// Save task
	if err := js.saveTask(task); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save task: %v", err)), nil
	}

	// Update daily log
	js.updateDailyLog(taskID, entry)

	return mcp.NewToolResultText(fmt.Sprintf("Updated task %s status to %s", taskID, status)), nil
}

// Placeholder implementations for remaining methods
func (js *JournalService) GetDailyLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("Daily log feature coming soon"), nil
}

func (js *JournalService) GetWeeklyLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("Weekly log feature coming soon"), nil
}

func (js *JournalService) CreateOneOnOne(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("1-on-1 feature coming soon"), nil
}

func (js *JournalService) GetOneOnOneHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("1-on-1 history feature coming soon"), nil
}

func (js *JournalService) SearchEntries(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("Search feature coming soon"), nil
}

func (js *JournalService) ExportData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("Export feature coming soon"), nil
}

// Helper methods
func (js *JournalService) saveTask(task *Task) error {
	filePath := filepath.Join(js.dataDir, "tasks", task.ID+".json")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (js *JournalService) loadTask(taskID string) (*Task, error) {
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

func (js *JournalService) loadAllTasks() ([]*Task, error) {
	tasksDir := filepath.Join(js.dataDir, "tasks")
	files, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			taskID := strings.TrimSuffix(file.Name(), ".json")
			if task, err := js.loadTask(taskID); err == nil {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

func (js *JournalService) filterTasks(tasks []*Task, filters map[string]interface{}) []*Task {
	var filtered []*Task

	for _, task := range tasks {
		include := true

		// Filter by status
		if status, exists := filters["status"].(string); exists && task.Status != status {
			include = false
		}

		// Filter by type
		if taskType, exists := filters["type"].(string); exists && task.Type != taskType {
			include = false
		}

		// Filter by tags
		if tagsRaw, exists := filters["tags"].([]interface{}); exists {
			hasTag := false
			for _, tagRaw := range tagsRaw {
				if tag, ok := tagRaw.(string); ok {
					for _, taskTag := range task.Tags {
						if taskTag == tag {
							hasTag = true
							break
						}
					}
					if hasTag {
						break
					}
				}
			}
			if !hasTag {
				include = false
			}
		}

		// TODO: Add date filtering

		if include {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

func (js *JournalService) formatTaskAsMarkdown(task *Task) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s: %s\n", task.ID, task.Title))
	md.WriteString(fmt.Sprintf("**Type:** %s | **Status:** %s", task.Type, task.Status))
	if task.Priority != "" {
		md.WriteString(fmt.Sprintf(" | **Priority:** %s", task.Priority))
	}
	md.WriteString("\n")

	if len(task.Tags) > 0 {
		md.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(task.Tags, ", ")))
	}

	if task.IssueURL != "" {
		md.WriteString(fmt.Sprintf("**Issue:** [%s](%s)\n", task.IssueID, task.IssueURL))
	}

	md.WriteString(fmt.Sprintf("**Created:** %s | **Updated:** %s\n\n",
		task.Created.Format("2006-01-02 15:04"),
		task.Updated.Format("2006-01-02 15:04")))

	// Group entries by date
	entriesByDate := make(map[string][]Entry)
	for _, entry := range task.Entries {
		date := entry.Timestamp.Format("2006-01-02")
		entriesByDate[date] = append(entriesByDate[date], entry)
	}

	// Sort dates
	var dates []string
	for date := range entriesByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Format entries by date
	for _, date := range dates {
		md.WriteString(fmt.Sprintf("## %s\n", date))
		entries := entriesByDate[date]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})

		for _, entry := range entries {
			md.WriteString(fmt.Sprintf("### %s\n", entry.Timestamp.Format("15:04")))
			md.WriteString(fmt.Sprintf("%s\n\n", entry.Content))
		}
	}

	return md.String()
}

func (js *JournalService) updateDailyLog(taskID string, entry Entry) {
	// TODO: Implement daily log updating
	// This would update a daily summary file with references to task entries
}

func generateEntryID() string {
	return fmt.Sprintf("entry_%d", time.Now().UnixNano())
}
