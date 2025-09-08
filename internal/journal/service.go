package journal

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
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
	backupPath := request.GetString("backup_path", "")
	includeConfig := request.GetString("include_config", "true") == "true"
	compressionLevel := request.GetString("compression", "default")

	if backupPath == "" {
		// Generate default backup path
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		backupPath = filepath.Join(js.dataDir, "backups", fmt.Sprintf("journal-backup-%s.zip", timestamp))
	}

	// Ensure backup directory exists
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create backup directory: %v", err)), nil
	}

	// Create ZIP file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create backup file: %v", err)), nil
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var filesBackup int
	var totalSize int64

	// Backup tasks
	tasksDir := filepath.Join(js.dataDir, "tasks")
	if err := js.addDirectoryToZip(zipWriter, tasksDir, "tasks", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup tasks: %v", err)), nil
	}

	// Backup daily logs
	dailyDir := filepath.Join(js.dataDir, "daily")
	if err := js.addDirectoryToZip(zipWriter, dailyDir, "daily", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup daily logs: %v", err)), nil
	}

	// Backup weekly logs
	weeklyDir := filepath.Join(js.dataDir, "weekly")
	if err := js.addDirectoryToZip(zipWriter, weeklyDir, "weekly", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup weekly logs: %v", err)), nil
	}

	// Backup one-on-ones
	oneOnOneDir := filepath.Join(js.dataDir, "one-on-ones")
	if err := js.addDirectoryToZip(zipWriter, oneOnOneDir, "one-on-ones", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup one-on-ones: %v", err)), nil
	}

	// Backup configuration if requested
	if includeConfig {
		configPath := filepath.Join(js.dataDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			if err := js.addFileToZip(zipWriter, configPath, "config.yaml", &totalSize); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to backup config: %v", err)), nil
			}
			filesBackup++
		}
	}

	// Close zip writer to finalize
	zipWriter.Close()
	
	// Get final file size
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get backup file info: %v", err)), nil
	}
	
	result := BackupResult{
		BackupPath:  backupPath,
		Size:        fileInfo.Size(),
		FilesBackup: filesBackup,
		CreatedAt:   time.Now(),
		Summary:     fmt.Sprintf("Backup completed successfully: %d files (%d bytes)", filesBackup, fileInfo.Size()),
	}

	response, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(response)), nil
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

func (js *Service) formatTaskAsMarkdown(task *Task) string {
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
