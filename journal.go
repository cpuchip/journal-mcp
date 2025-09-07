package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

type ImportResult struct {
	TasksCreated     int      `json:"tasks_created"`
	EntriesAdded     int      `json:"entries_added"`
	DuplicatesSkipped int      `json:"duplicates_skipped"`
	Warnings         []string `json:"warnings,omitempty"`
	Summary          string   `json:"summary"`
}

type TaskRecommendation struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Rationale   string  `json:"rationale"`
	Priority    string  `json:"priority"`
	Confidence  float64 `json:"confidence"`
	SuggestedTags []string `json:"suggested_tags,omitempty"`
}

type RecommendationsResult struct {
	Recommendations []TaskRecommendation `json:"recommendations"`
	AnalysisMetrics map[string]interface{} `json:"analysis_metrics"`
	Summary         string                 `json:"summary"`
}

type AnalyticsReport struct {
	ReportType      string                 `json:"report_type"`
	TimePeriod      string                 `json:"time_period"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Summary         string                 `json:"summary"`
	TaskMetrics     TaskMetrics            `json:"task_metrics"`
	ProductivityMetrics ProductivityMetrics `json:"productivity_metrics"`
	PatternAnalysis PatternAnalysis        `json:"pattern_analysis"`
	Trends          []Trend                `json:"trends,omitempty"`
	Insights        []string               `json:"insights"`
}

type TaskMetrics struct {
	TotalTasks       int                 `json:"total_tasks"`
	ByStatus         map[string]int      `json:"by_status"`
	ByType           map[string]int      `json:"by_type"`
	ByPriority       map[string]int      `json:"by_priority"`
	CompletionRate   float64             `json:"completion_rate"`
	AverageEntries   float64             `json:"average_entries_per_task"`
	TotalEntries     int                 `json:"total_entries"`
}

type ProductivityMetrics struct {
	TasksCompletedPeriod  int     `json:"tasks_completed_period"`
	EntriesAddedPeriod    int     `json:"entries_added_period"`
	AverageTaskDuration   float64 `json:"average_task_duration_days"`
	MostProductiveType    string  `json:"most_productive_type"`
	ProductivityScore     float64 `json:"productivity_score"`
}

type PatternAnalysis struct {
	MostFrequentType      string                `json:"most_frequent_type"`
	CommonTags            []string              `json:"common_tags"`
	WorkPatterns          map[string]int        `json:"work_patterns"`
	TimeToCompletion      map[string]float64    `json:"time_to_completion_by_type"`
}

type Trend struct {
	Metric    string  `json:"metric"`
	Direction string  `json:"direction"` // "up", "down", "stable"
	Change    float64 `json:"change"`
	Period    string  `json:"period"`
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

	// Apply pagination
	limit := 50 // default
	if limitStr := request.GetString("limit", ""); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 200 {
				parsedLimit = 200 // max limit
			}
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := request.GetString("offset", ""); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	totalTasks := len(filtered)
	startIndex := offset
	endIndex := offset + limit
	
	if startIndex >= totalTasks {
		startIndex = totalTasks
		endIndex = totalTasks
	} else if endIndex > totalTasks {
		endIndex = totalTasks
	}

	paginatedTasks := filtered[startIndex:endIndex]

	// Format as list
	var result strings.Builder
	result.WriteString(fmt.Sprintf("# Task List (showing %d-%d of %d total)\n\n", 
		startIndex+1, endIndex, totalTasks))

	if len(paginatedTasks) == 0 {
		result.WriteString("No tasks found matching the criteria.")
		return mcp.NewToolResultText(result.String()), nil
	}

	for _, task := range paginatedTasks {
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

// Remaining MCP tool implementations
func (js *JournalService) GetDailyLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date, err := request.RequireString("date")
	if err != nil {
		return mcp.NewToolResultError("date is required (YYYY-MM-DD format)"), nil
	}

	// Validate date format
	if validationErr := js.validateDateFormat(date, "date"); validationErr != nil {
		return mcp.NewToolResultError(validationErr.Error()), nil
	}

	// Load daily activity file if it exists
	dailyPath := filepath.Join(js.dataDir, "daily", date+".json")
	var dailyActivity DailyActivity
	
	if data, err := os.ReadFile(dailyPath); err == nil {
		json.Unmarshal(data, &dailyActivity)
	} else {
		// Create new daily activity by scanning all tasks for entries on this date
		dailyActivity = DailyActivity{
			Date:  date,
			Tasks: make(map[string][]Entry),
		}
		
		// Load all tasks and filter entries for this date
		tasks, err := js.loadAllTasks()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
		}
		
		for _, task := range tasks {
			var dayEntries []Entry
			for _, entry := range task.Entries {
				if entry.Timestamp.Format("2006-01-02") == date {
					dayEntries = append(dayEntries, entry)
				}
			}
			if len(dayEntries) > 0 {
				dailyActivity.Tasks[task.ID] = dayEntries
			}
		}
		
		// Save the daily activity for future reference
		js.saveDailyActivity(&dailyActivity)
	}

	// Format as markdown
	markdown := js.formatDailyLogAsMarkdown(&dailyActivity)
	
	return mcp.NewToolResultText(markdown), nil
}

func (js *JournalService) GetWeeklyLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	weekStart, err := request.RequireString("week_start")
	if err != nil {
		return mcp.NewToolResultError("week_start is required (YYYY-MM-DD format)"), nil
	}

	// Validate date format
	if validationErr := js.validateDateFormat(weekStart, "week_start"); validationErr != nil {
		return mcp.NewToolResultError(validationErr.Error()), nil
	}
	
	startDate, _ := time.Parse("2006-01-02", weekStart) // Safe to parse since validation passed

	var weeklyMarkdown strings.Builder
	weeklyMarkdown.WriteString(fmt.Sprintf("# Weekly Log: %s to %s\n\n", 
		startDate.Format("2006-01-02"), 
		startDate.AddDate(0, 0, 6).Format("2006-01-02")))

	totalEntries := 0
	tasksWorked := make(map[string]bool)

	// Aggregate daily logs for 7 days
	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		
		// Load daily activity file if it exists
		dailyPath := filepath.Join(js.dataDir, "daily", dateStr+".json")
		var dailyActivity DailyActivity
		
		if data, err := os.ReadFile(dailyPath); err == nil {
			if json.Unmarshal(data, &dailyActivity) == nil {
				for taskID, entries := range dailyActivity.Tasks {
					tasksWorked[taskID] = true
					totalEntries += len(entries)
				}
			}
		} else {
			// Create daily activity by scanning tasks for this date
			tasks, err := js.loadAllTasks()
			if err == nil {
				dailyActivity = DailyActivity{
					Date:  dateStr,
					Tasks: make(map[string][]Entry),
				}
				
				for _, task := range tasks {
					var dayEntries []Entry
					for _, entry := range task.Entries {
						if entry.Timestamp.Format("2006-01-02") == dateStr {
							dayEntries = append(dayEntries, entry)
						}
					}
					if len(dayEntries) > 0 {
						dailyActivity.Tasks[task.ID] = dayEntries
						tasksWorked[task.ID] = true
						totalEntries += len(dayEntries)
					}
				}
			}
		}
		
		weeklyMarkdown.WriteString(fmt.Sprintf("## %s (%s)\n", 
			dateStr, currentDate.Format("Monday")))
		
		if len(dailyActivity.Tasks) == 0 {
			weeklyMarkdown.WriteString("_No activity_\n\n")
		} else {
			for taskID, entries := range dailyActivity.Tasks {
				if task, err := js.loadTask(taskID); err == nil {
					weeklyMarkdown.WriteString(fmt.Sprintf("### %s: %s\n", taskID, task.Title))
				} else {
					weeklyMarkdown.WriteString(fmt.Sprintf("### %s\n", taskID))
				}
				
				for _, entry := range entries {
					weeklyMarkdown.WriteString(fmt.Sprintf("- %s: %s\n", 
						entry.Timestamp.Format("15:04"), entry.Content))
				}
				weeklyMarkdown.WriteString("\n")
			}
		}
	}
	
	// Add weekly summary
	weeklyMarkdown.WriteString("## Weekly Summary\n")
	weeklyMarkdown.WriteString(fmt.Sprintf("- **Total entries:** %d\n", totalEntries))
	weeklyMarkdown.WriteString(fmt.Sprintf("- **Tasks worked on:** %d\n", len(tasksWorked)))
	if len(tasksWorked) > 0 {
		weeklyMarkdown.WriteString("- **Tasks:** ")
		var taskIDs []string
		for taskID := range tasksWorked {
			taskIDs = append(taskIDs, taskID)
		}
		sort.Strings(taskIDs)
		weeklyMarkdown.WriteString(strings.Join(taskIDs, ", ") + "\n")
	}

	return mcp.NewToolResultText(weeklyMarkdown.String()), nil
}

func (js *JournalService) CreateOneOnOne(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	date, err := request.RequireString("date")
	if err != nil {
		return mcp.NewToolResultError("date is required (YYYY-MM-DD format)"), nil
	}

	// Validate date format
	if validationErr := js.validateDateFormat(date, "date"); validationErr != nil {
		return mcp.NewToolResultError(validationErr.Error()), nil
	}

	oneOnOne := OneOnOne{
		Date:    date,
		Created: time.Now(),
	}

	// Parse optional arrays
	if insights := request.GetStringSlice("insights", nil); insights != nil {
		oneOnOne.Insights = insights
	}
	
	if todos := request.GetStringSlice("todos", nil); todos != nil {
		oneOnOne.Todos = todos
	}
	
	if feedback := request.GetStringSlice("feedback", nil); feedback != nil {
		oneOnOne.Feedback = feedback
	}

	// Parse optional notes
	if notes := request.GetString("notes", ""); notes != "" {
		oneOnOne.Notes = notes
	}

	// Save to file
	filePath := filepath.Join(js.dataDir, "one-on-ones", date+".json")
	data, err := json.MarshalIndent(oneOnOne, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize one-on-one: %v", err)), nil
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save one-on-one: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created one-on-one meeting notes for %s", date)), nil
}

func (js *JournalService) GetOneOnOneHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := 10 // default
	if limitStr := request.GetString("limit", ""); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Read all one-on-one files
	oneOnOnesDir := filepath.Join(js.dataDir, "one-on-ones")
	files, err := os.ReadDir(oneOnOnesDir)
	if err != nil {
		// Directory might not exist yet
		return mcp.NewToolResultText("# One-on-One History\n\nNo one-on-one meetings recorded yet."), nil
	}

	var oneOnOnes []OneOnOne
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(oneOnOnesDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			
			var oneOnOne OneOnOne
			if err := json.Unmarshal(data, &oneOnOne); err == nil {
				oneOnOnes = append(oneOnOnes, oneOnOne)
			}
		}
	}

	// Sort by date (most recent first)
	sort.Slice(oneOnOnes, func(i, j int) bool {
		return oneOnOnes[i].Date > oneOnOnes[j].Date
	})

	// Apply limit
	if len(oneOnOnes) > limit {
		oneOnOnes = oneOnOnes[:limit]
	}

	// Format as markdown
	var markdown strings.Builder
	markdown.WriteString("# One-on-One History\n\n")
	
	if len(oneOnOnes) == 0 {
		markdown.WriteString("No one-on-one meetings recorded yet.")
		return mcp.NewToolResultText(markdown.String()), nil
	}

	for _, meeting := range oneOnOnes {
		markdown.WriteString(fmt.Sprintf("## %s\n", meeting.Date))
		
		if len(meeting.Insights) > 0 {
			markdown.WriteString("**Insights:**\n")
			for _, insight := range meeting.Insights {
				markdown.WriteString(fmt.Sprintf("- %s\n", insight))
			}
			markdown.WriteString("\n")
		}
		
		if len(meeting.Todos) > 0 {
			markdown.WriteString("**Action Items:**\n")
			for _, todo := range meeting.Todos {
				markdown.WriteString(fmt.Sprintf("- [ ] %s\n", todo))
			}
			markdown.WriteString("\n")
		}
		
		if len(meeting.Feedback) > 0 {
			markdown.WriteString("**Feedback:**\n")
			for _, fb := range meeting.Feedback {
				markdown.WriteString(fmt.Sprintf("- %s\n", fb))
			}
			markdown.WriteString("\n")
		}
		
		if meeting.Notes != "" {
			markdown.WriteString("**Notes:**\n")
			markdown.WriteString(meeting.Notes + "\n\n")
		}
		
		markdown.WriteString("---\n\n")
	}

	return mcp.NewToolResultText(markdown.String()), nil
}

func (js *JournalService) SearchEntries(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil
	}

	query = strings.ToLower(query)
	
	// Optional filters
	taskType := request.GetString("task_type", "")
	dateFrom := request.GetString("date_from", "")
	dateTo := request.GetString("date_to", "")

	// Parse dates safely (invalid dates are ignored with warning in logs)
	fromTime := js.parseDateSafely(dateFrom)
	toTime := js.parseDateSafely(dateTo)
	if !toTime.IsZero() {
		toTime = toTime.AddDate(0, 0, 1) // Include the entire day
	}

	// Search through all tasks
	tasks, err := js.loadAllTasks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
	}

	type SearchResult struct {
		TaskID    string
		TaskTitle string
		Entry     Entry
		Context   string
	}

	var results []SearchResult

	for _, task := range tasks {
		// Filter by task type if specified
		if taskType != "" && task.Type != taskType {
			continue
		}

		// Search in task title and entries
		taskMatches := strings.Contains(strings.ToLower(task.Title), query)
		
		for _, entry := range task.Entries {
			// Filter by date range if specified
			if !fromTime.IsZero() && entry.Timestamp.Before(fromTime) {
				continue
			}
			if !toTime.IsZero() && entry.Timestamp.After(toTime) {
				continue
			}

			entryMatches := strings.Contains(strings.ToLower(entry.Content), query)
			
			if taskMatches || entryMatches {
				context := "task"
				if entryMatches {
					context = "entry"
				}
				if taskMatches && entryMatches {
					context = "both"
				}
				
				results = append(results, SearchResult{
					TaskID:    task.ID,
					TaskTitle: task.Title,
					Entry:     entry,
					Context:   context,
				})
			}
		}
	}

	// Search through one-on-ones
	oneOnOnesDir := filepath.Join(js.dataDir, "one-on-ones")
	if files, err := os.ReadDir(oneOnOnesDir); err == nil {
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			
			filePath := filepath.Join(oneOnOnesDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			
			var oneOnOne OneOnOne
			if json.Unmarshal(data, &oneOnOne) != nil {
				continue
			}
			
			// Filter by date if specified
			meetingDate, _ := time.Parse("2006-01-02", oneOnOne.Date)
			if !fromTime.IsZero() && meetingDate.Before(fromTime) {
				continue
			}
			if !toTime.IsZero() && meetingDate.After(toTime) {
				continue
			}
			
			// Search in one-on-one content
			searchText := strings.ToLower(oneOnOne.Notes + " " + strings.Join(oneOnOne.Insights, " ") + " " + strings.Join(oneOnOne.Todos, " ") + " " + strings.Join(oneOnOne.Feedback, " "))
			if strings.Contains(searchText, query) {
				results = append(results, SearchResult{
					TaskID:    "one-on-one",
					TaskTitle: fmt.Sprintf("One-on-One: %s", oneOnOne.Date),
					Entry: Entry{
						Timestamp: meetingDate,
						Content:   oneOnOne.Notes,
						Type:      "one-on-one",
					},
					Context: "one-on-one",
				})
			}
		}
	}

	// Sort results by relevance (timestamp descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Entry.Timestamp.After(results[j].Entry.Timestamp)
	})

	// Format results
	var markdown strings.Builder
	markdown.WriteString(fmt.Sprintf("# Search Results for \"%s\"\n\n", query))
	
	if len(results) == 0 {
		markdown.WriteString("No matching entries found.")
		return mcp.NewToolResultText(markdown.String()), nil
	}
	
	markdown.WriteString(fmt.Sprintf("Found %d matching entries:\n\n", len(results)))

	for _, result := range results {
		markdown.WriteString(fmt.Sprintf("## %s: %s\n", result.TaskID, result.TaskTitle))
		markdown.WriteString(fmt.Sprintf("**Date:** %s | **Context:** %s\n\n", 
			result.Entry.Timestamp.Format("2006-01-02 15:04"), result.Context))
		
		// Highlight the matching content (simple approach)
		content := result.Entry.Content
		if len(content) > 200 {
			// Find the query position and show context around it
			lowerContent := strings.ToLower(content)
			queryPos := strings.Index(lowerContent, query)
			if queryPos >= 0 {
				start := max(0, queryPos-50)
				end := min(len(content), queryPos+len(query)+100)
				content = "..." + content[start:end] + "..."
			} else {
				content = content[:200] + "..."
			}
		}
		
		markdown.WriteString(content + "\n\n---\n\n")
	}

	return mcp.NewToolResultText(markdown.String()), nil
}

func (js *JournalService) ExportData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format, err := request.RequireString("format")
	if err != nil {
		return mcp.NewToolResultError("format is required (json|markdown|csv)"), nil
	}

	// Validate format
	if format != "json" && format != "markdown" && format != "csv" {
		return mcp.NewToolResultError("Invalid format. Must be: json, markdown, csv"), nil
	}

	// Optional filters
	dateFrom := request.GetString("date_from", "")
	dateTo := request.GetString("date_to", "")
	taskFilter := request.GetString("task_filter", "")

	// Parse dates safely (invalid dates are ignored)
	fromTime := js.parseDateSafely(dateFrom)
	toTime := js.parseDateSafely(dateTo)
	if !toTime.IsZero() {
		toTime = toTime.AddDate(0, 0, 1) // Include the entire day
	}

	// Load and filter tasks
	tasks, err := js.loadAllTasks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
	}

	var filteredTasks []*Task
	for _, task := range tasks {
		if taskFilter != "" && task.Type != taskFilter {
			continue
		}
		
		// Filter entries by date if specified
		var filteredEntries []Entry
		for _, entry := range task.Entries {
			if !fromTime.IsZero() && entry.Timestamp.Before(fromTime) {
				continue
			}
			if !toTime.IsZero() && entry.Timestamp.After(toTime) {
				continue
			}
			filteredEntries = append(filteredEntries, entry)
		}
		
		if len(filteredEntries) > 0 || (fromTime.IsZero() && toTime.IsZero()) {
			filteredTask := *task
			if !fromTime.IsZero() || !toTime.IsZero() {
				filteredTask.Entries = filteredEntries
			}
			filteredTasks = append(filteredTasks, &filteredTask)
		}
	}

	// Load one-on-ones if in date range
	var oneOnOnes []OneOnOne
	oneOnOnesDir := filepath.Join(js.dataDir, "one-on-ones")
	if files, err := os.ReadDir(oneOnOnesDir); err == nil {
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			
			filePath := filepath.Join(oneOnOnesDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			
			var oneOnOne OneOnOne
			if json.Unmarshal(data, &oneOnOne) != nil {
				continue
			}
			
			// Filter by date if specified
			meetingDate, _ := time.Parse("2006-01-02", oneOnOne.Date)
			if !fromTime.IsZero() && meetingDate.Before(fromTime) {
				continue
			}
			if !toTime.IsZero() && meetingDate.After(toTime) {
				continue
			}
			
			oneOnOnes = append(oneOnOnes, oneOnOne)
		}
	}

	// Export based on format
	switch format {
	case "json":
		exportData := map[string]interface{}{
			"tasks":      filteredTasks,
			"one_on_ones": oneOnOnes,
			"exported_at": time.Now().Format(time.RFC3339),
		}
		
		jsonData, err := json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
		}
		
		return mcp.NewToolResultText(string(jsonData)), nil

	case "markdown":
		var md strings.Builder
		md.WriteString("# Journal Export\n\n")
		md.WriteString(fmt.Sprintf("Exported on: %s\n\n", time.Now().Format("2006-01-02 15:04")))
		
		if len(filteredTasks) > 0 {
			md.WriteString("## Tasks\n\n")
			for _, task := range filteredTasks {
				md.WriteString(js.formatTaskAsMarkdown(task))
				md.WriteString("\n---\n\n")
			}
		}
		
		if len(oneOnOnes) > 0 {
			md.WriteString("## One-on-One Meetings\n\n")
			for _, meeting := range oneOnOnes {
				md.WriteString(fmt.Sprintf("### %s\n", meeting.Date))
				if len(meeting.Insights) > 0 {
					md.WriteString("**Insights:**\n")
					for _, insight := range meeting.Insights {
						md.WriteString(fmt.Sprintf("- %s\n", insight))
					}
				}
				if len(meeting.Todos) > 0 {
					md.WriteString("**Action Items:**\n")
					for _, todo := range meeting.Todos {
						md.WriteString(fmt.Sprintf("- [ ] %s\n", todo))
					}
				}
				if meeting.Notes != "" {
					md.WriteString("**Notes:**\n" + meeting.Notes + "\n")
				}
				md.WriteString("\n")
			}
		}
		
		return mcp.NewToolResultText(md.String()), nil

	case "csv":
		var csv strings.Builder
		csv.WriteString("Type,Date,Time,Task_ID,Task_Title,Content,Entry_Type\n")
		
		for _, task := range filteredTasks {
			for _, entry := range task.Entries {
				csv.WriteString(fmt.Sprintf("task,%s,%s,%s,\"%s\",\"%s\",%s\n",
					entry.Timestamp.Format("2006-01-02"),
					entry.Timestamp.Format("15:04"),
					task.ID,
					strings.ReplaceAll(task.Title, "\"", "\"\""),
					strings.ReplaceAll(entry.Content, "\"", "\"\""),
					entry.Type))
			}
		}
		
		for _, meeting := range oneOnOnes {
			content := meeting.Notes
			if len(meeting.Insights) > 0 {
				content += " | Insights: " + strings.Join(meeting.Insights, "; ")
			}
			if len(meeting.Todos) > 0 {
				content += " | Todos: " + strings.Join(meeting.Todos, "; ")
			}
			
			csv.WriteString(fmt.Sprintf("one-on-one,%s,00:00,one-on-one,\"One-on-One Meeting\",\"%s\",meeting\n",
				meeting.Date,
				strings.ReplaceAll(content, "\"", "\"\"")))
		}
		
		return mcp.NewToolResultText(csv.String()), nil
	}

	return mcp.NewToolResultError("Unsupported format"), nil
}

func (js *JournalService) ImportData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	if strings.TrimSpace(content) == "" {
		return mcp.NewToolResultError("content cannot be empty"), nil
	}

	format, err := request.RequireString("format")
	if err != nil {
		return mcp.NewToolResultError("format is required"), nil
	}

	taskPrefix := request.GetString("task_prefix", "IMPORT")
	defaultType := request.GetString("default_type", "personal")

	// Validate format
	validFormats := map[string]bool{"txt": true, "markdown": true, "json": true, "csv": true}
	if !validFormats[format] {
		return mcp.NewToolResultError("format must be one of: txt, markdown, json, csv"), nil
	}

	// Validate default type
	validTypes := map[string]bool{"work": true, "learning": true, "personal": true, "investigation": true}
	if !validTypes[defaultType] {
		return mcp.NewToolResultError("default_type must be one of: work, learning, personal, investigation"), nil
	}

	var result ImportResult
	var warnings []string

	switch format {
	case "txt":
		result, warnings = js.importFromPlainText(content, taskPrefix, defaultType)
	case "markdown":
		result, warnings = js.importFromMarkdown(content, taskPrefix, defaultType)
	case "json":
		result, warnings = js.importFromJSON(content, taskPrefix, defaultType)
	case "csv":
		result, warnings = js.importFromCSV(content, taskPrefix, defaultType)
	}

	result.Warnings = warnings
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (js *JournalService) GetTaskRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskType := request.GetString("task_type", "")
	focusArea := request.GetString("focus_area", "productivity")
	limitStr := request.GetString("limit", "5")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 20 {
		limit = 5
	}
	
	// Validate focus area
	validFocus := map[string]bool{"productivity": true, "learning": true, "completion": true, "priority": true}
	if !validFocus[focusArea] {
		return mcp.NewToolResultError("focus_area must be one of: productivity, learning, completion, priority"), nil
	}
	
	// Load all tasks for analysis
	tasks, err := js.loadAllTasks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
	}
	
	// Analyze patterns and generate recommendations
	recommendations := js.analyzeAndRecommend(tasks, taskType, focusArea, limit)
	
	result := RecommendationsResult{
		Recommendations: recommendations,
		AnalysisMetrics: js.calculateAnalysisMetrics(tasks),
		Summary:         fmt.Sprintf("Generated %d recommendations based on %s analysis of %d tasks", 
			len(recommendations), focusArea, len(tasks)),
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (js *JournalService) GetAnalyticsReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reportType := request.GetString("report_type", "overview")
	timePeriod := request.GetString("time_period", "month")
	taskType := request.GetString("task_type", "")
	
	// Validate parameters
	validReportTypes := map[string]bool{"overview": true, "productivity": true, "patterns": true, "trends": true}
	if !validReportTypes[reportType] {
		return mcp.NewToolResultError("report_type must be one of: overview, productivity, patterns, trends"), nil
	}
	
	validTimePeriods := map[string]bool{"week": true, "month": true, "quarter": true, "year": true, "all": true}
	if !validTimePeriods[timePeriod] {
		return mcp.NewToolResultError("time_period must be one of: week, month, quarter, year, all"), nil
	}
	
	// Load all tasks for analysis
	allTasks, err := js.loadAllTasks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
	}
	
	// Filter tasks by time period and type
	filteredTasks := js.filterTasksByTimePeriod(allTasks, timePeriod)
	if taskType != "" {
		filteredTasks = js.getTasksByType(filteredTasks, taskType)
	}
	
	// Generate analytics report
	report := js.generateAnalyticsReport(filteredTasks, reportType, timePeriod)
	
	resultJSON, _ := json.MarshalIndent(report, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
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

		// Filter by date range (using Updated timestamp)
		if dateFromStr, exists := filters["date_from"].(string); exists && dateFromStr != "" {
			dateFrom := js.parseDateSafely(dateFromStr)
			if !dateFrom.IsZero() && task.Updated.Before(dateFrom) {
				include = false
			}
		}

		if dateToStr, exists := filters["date_to"].(string); exists && dateToStr != "" {
			dateTo := js.parseDateSafely(dateToStr)
			if !dateTo.IsZero() {
				// Add one day to include the entire end date
				endOfDay := dateTo.Add(24 * time.Hour)
				if task.Updated.After(endOfDay) {
					include = false
				}
			}
		}

		if include {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// validateDateFormat validates a date string and returns a user-friendly error
func (js *JournalService) validateDateFormat(dateStr, fieldName string) error {
	if dateStr == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("Invalid %s format. Expected YYYY-MM-DD (e.g., 2025-01-15), got: %s", fieldName, dateStr)
	}
	
	return nil
}

// parseDateSafely parses a date string and returns zero time if invalid
func (js *JournalService) parseDateSafely(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	
	if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
		return parsed
	}
	
	return time.Time{}
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
	date := entry.Timestamp.Format("2006-01-02")
	dailyPath := filepath.Join(js.dataDir, "daily", date+".json")
	
	var dailyActivity DailyActivity
	
	// Load existing daily activity or create new one
	if data, err := os.ReadFile(dailyPath); err == nil {
		json.Unmarshal(data, &dailyActivity)
	} else {
		dailyActivity = DailyActivity{
			Date:  date,
			Tasks: make(map[string][]Entry),
		}
	}
	
	// Add entry to the task's entries for this day
	dailyActivity.Tasks[taskID] = append(dailyActivity.Tasks[taskID], entry)
	
	// Save updated daily activity
	js.saveDailyActivity(&dailyActivity)
}

func (js *JournalService) saveDailyActivity(activity *DailyActivity) error {
	filePath := filepath.Join(js.dataDir, "daily", activity.Date+".json")
	data, err := json.MarshalIndent(activity, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (js *JournalService) formatDailyLogAsMarkdown(activity *DailyActivity) string {
	var md strings.Builder
	
	md.WriteString(fmt.Sprintf("# Daily Log: %s\n\n", activity.Date))
	
	if len(activity.Tasks) == 0 {
		md.WriteString("No activity recorded for this date.")
		return md.String()
	}
	
	// Sort task IDs
	var taskIDs []string
	for taskID := range activity.Tasks {
		taskIDs = append(taskIDs, taskID)
	}
	sort.Strings(taskIDs)
	
	for _, taskID := range taskIDs {
		entries := activity.Tasks[taskID]
		
		// Get task title for better display
		if task, err := js.loadTask(taskID); err == nil {
			md.WriteString(fmt.Sprintf("## %s: %s\n", taskID, task.Title))
		} else {
			md.WriteString(fmt.Sprintf("## %s\n", taskID))
		}
		
		// Sort entries by time
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

func generateEntryID() string {
	return fmt.Sprintf("entry_%d", time.Now().UnixNano())
}

// Helper functions for search
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Import helper functions
func (js *JournalService) importFromPlainText(content, taskPrefix, defaultType string) (ImportResult, []string) {
	var result ImportResult
	var warnings []string
	
	lines := strings.Split(content, "\n")
	currentTaskID := fmt.Sprintf("%s-%d", taskPrefix, time.Now().Unix())
	currentTask := &Task{
		ID:      currentTaskID,
		Title:   "Imported from plain text",
		Type:    defaultType,
		Status:  "active",
		Created: time.Now(),
		Updated: time.Now(),
		Entries: []Entry{},
	}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Try to parse date/time patterns
		timestamp := time.Now()
		content := line
		
		// Simple date pattern detection (YYYY-MM-DD or MM/DD/YYYY)
		if matched, parsedTime := js.extractTimestamp(line); matched {
			timestamp = parsedTime
			// Remove timestamp from content
			content = strings.TrimSpace(strings.Replace(line, parsedTime.Format("2006-01-02"), "", 1))
			content = strings.TrimSpace(strings.Replace(content, parsedTime.Format("01/02/2006"), "", 1))
		}
		
		if content != "" {
			entry := Entry{
				ID:        generateEntryID(),
				Timestamp: timestamp,
				Content:   content,
				Type:      "imported",
			}
			
			currentTask.Entries = append(currentTask.Entries, entry)
			result.EntriesAdded++
		}
	}
	
	if len(currentTask.Entries) > 0 {
		if err := js.saveTask(currentTask); err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to save task: %v", err))
		} else {
			result.TasksCreated++
		}
	}
	
	result.Summary = fmt.Sprintf("Imported %d entries into %d task(s) from plain text", result.EntriesAdded, result.TasksCreated)
	return result, warnings
}

func (js *JournalService) importFromMarkdown(content, taskPrefix, defaultType string) (ImportResult, []string) {
	var result ImportResult
	var warnings []string
	
	lines := strings.Split(content, "\n")
	var currentTask *Task
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check for headers (new tasks)
		if strings.HasPrefix(line, "#") {
			// Save previous task if exists
			if currentTask != nil && len(currentTask.Entries) > 0 {
				if err := js.saveTask(currentTask); err != nil {
					warnings = append(warnings, fmt.Sprintf("Failed to save task %s: %v", currentTask.ID, err))
				} else {
					result.TasksCreated++
				}
			}
			
			// Create new task from header
			title := strings.TrimSpace(strings.TrimLeft(line, "#"))
			taskID := fmt.Sprintf("%s-%d-%d", taskPrefix, time.Now().Unix(), lineNum)
			currentTask = &Task{
				ID:      taskID,
				Title:   title,
				Type:    defaultType,
				Status:  "active",
				Created: time.Now(),
				Updated: time.Now(),
				Entries: []Entry{},
			}
		} else if currentTask != nil {
			// Add content as entry
			timestamp := time.Now()
			if matched, parsedTime := js.extractTimestamp(line); matched {
				timestamp = parsedTime
			}
			
			entry := Entry{
				ID:        generateEntryID(),
				Timestamp: timestamp,
				Content:   line,
				Type:      "imported",
			}
			
			currentTask.Entries = append(currentTask.Entries, entry)
			result.EntriesAdded++
		} else {
			// No current task, create default one
			taskID := fmt.Sprintf("%s-%d", taskPrefix, time.Now().Unix())
			currentTask = &Task{
				ID:      taskID,
				Title:   "Imported from markdown",
				Type:    defaultType,
				Status:  "active",
				Created: time.Now(),
				Updated: time.Now(),
				Entries: []Entry{},
			}
			
			entry := Entry{
				ID:        generateEntryID(),
				Timestamp: time.Now(),
				Content:   line,
				Type:      "imported",
			}
			
			currentTask.Entries = append(currentTask.Entries, entry)
			result.EntriesAdded++
		}
	}
	
	// Save last task
	if currentTask != nil && len(currentTask.Entries) > 0 {
		if err := js.saveTask(currentTask); err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to save task %s: %v", currentTask.ID, err))
		} else {
			result.TasksCreated++
		}
	}
	
	result.Summary = fmt.Sprintf("Imported %d entries into %d task(s) from markdown", result.EntriesAdded, result.TasksCreated)
	return result, warnings
}

func (js *JournalService) importFromJSON(content, taskPrefix, defaultType string) (ImportResult, []string) {
	var result ImportResult
	var warnings []string
	
	var validTypes = map[string]bool{"work": true, "learning": true, "personal": true, "investigation": true}
	
	// Try to parse as our format first
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		warnings = append(warnings, fmt.Sprintf("Invalid JSON format: %v", err))
		result.Summary = "Failed to parse JSON"
		return result, warnings
	}
	
	// Check if it matches our export format
	if tasks, ok := data["tasks"].([]interface{}); ok {
		for _, taskData := range tasks {
			taskMap, ok := taskData.(map[string]interface{})
			if !ok {
				continue
			}
			
			// Parse task
			task := &Task{
				ID:      fmt.Sprintf("%s-%s", taskPrefix, taskMap["id"].(string)),
				Title:   taskMap["title"].(string),
				Type:    defaultType,
				Status:  "active",
				Created: time.Now(),
				Updated: time.Now(),
				Entries: []Entry{},
			}
			
			if taskType, ok := taskMap["type"].(string); ok && validTypes[taskType] {
				task.Type = taskType
			}
			
			// Parse entries
			if entries, ok := taskMap["entries"].([]interface{}); ok {
				for _, entryData := range entries {
					entryMap, ok := entryData.(map[string]interface{})
					if !ok {
						continue
					}
					
					timestamp := time.Now()
					if timestampStr, ok := entryMap["timestamp"].(string); ok {
						if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
							timestamp = parsedTime
						}
					}
					
					entry := Entry{
						ID:        generateEntryID(),
						Timestamp: timestamp,
						Content:   entryMap["content"].(string),
						Type:      "imported",
					}
					
					task.Entries = append(task.Entries, entry)
					result.EntriesAdded++
				}
			}
			
			if len(task.Entries) > 0 {
				if err := js.saveTask(task); err != nil {
					warnings = append(warnings, fmt.Sprintf("Failed to save task %s: %v", task.ID, err))
				} else {
					result.TasksCreated++
				}
			}
		}
	} else {
		// Generic JSON - treat as single task
		taskID := fmt.Sprintf("%s-%d", taskPrefix, time.Now().Unix())
		task := &Task{
			ID:      taskID,
			Title:   "Imported from JSON",
			Type:    defaultType,
			Status:  "active",
			Created: time.Now(),
			Updated: time.Now(),
			Entries: []Entry{},
		}
		
		entry := Entry{
			ID:        generateEntryID(),
			Timestamp: time.Now(),
			Content:   content,
			Type:      "imported",
		}
		
		task.Entries = append(task.Entries, entry)
		result.EntriesAdded++
		
		if err := js.saveTask(task); err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to save task: %v", err))
		} else {
			result.TasksCreated++
		}
	}
	
	result.Summary = fmt.Sprintf("Imported %d entries into %d task(s) from JSON", result.EntriesAdded, result.TasksCreated)
	return result, warnings
}

func (js *JournalService) importFromCSV(content, taskPrefix, defaultType string) (ImportResult, []string) {
	var result ImportResult
	var warnings []string
	
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		warnings = append(warnings, "CSV must have at least header and one data row")
		result.Summary = "Invalid CSV format"
		return result, warnings
	}
	
	// Parse header
	header := strings.Split(lines[0], ",")
	for i := range header {
		header[i] = strings.TrimSpace(strings.Trim(header[i], "\""))
	}
	
	// Find column indices
	titleCol, dateCol, contentCol := -1, -1, -1
	for i, col := range header {
		switch strings.ToLower(col) {
		case "title", "task", "name":
			titleCol = i
		case "date", "timestamp", "time":
			dateCol = i
		case "content", "description", "notes", "entry":
			contentCol = i
		}
	}
	
	if contentCol == -1 {
		warnings = append(warnings, "Could not find content/description column")
		result.Summary = "Invalid CSV format - missing content column"
		return result, warnings
	}
	
	taskMap := make(map[string]*Task)
	
	// Process data rows
	for lineNum, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		fields := js.parseCSVLine(line)
		if len(fields) <= max(titleCol, max(dateCol, contentCol)) {
			warnings = append(warnings, fmt.Sprintf("Row %d has insufficient columns", lineNum+2))
			continue
		}
		
		// Determine task ID
		taskID := fmt.Sprintf("%s-%d", taskPrefix, time.Now().Unix())
		if titleCol >= 0 && titleCol < len(fields) && strings.TrimSpace(fields[titleCol]) != "" {
			taskTitle := strings.TrimSpace(strings.Trim(fields[titleCol], "\""))
			taskID = fmt.Sprintf("%s-%s", taskPrefix, strings.ReplaceAll(taskTitle, " ", "-"))
		}
		
		// Get or create task
		if _, exists := taskMap[taskID]; !exists {
			title := "Imported from CSV"
			if titleCol >= 0 && titleCol < len(fields) {
				title = strings.TrimSpace(strings.Trim(fields[titleCol], "\""))
			}
			
			taskMap[taskID] = &Task{
				ID:      taskID,
				Title:   title,
				Type:    defaultType,
				Status:  "active",
				Created: time.Now(),
				Updated: time.Now(),
				Entries: []Entry{},
			}
		}
		
		// Parse timestamp
		timestamp := time.Now()
		if dateCol >= 0 && dateCol < len(fields) {
			dateStr := strings.TrimSpace(strings.Trim(fields[dateCol], "\""))
			if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
				timestamp = parsed
			} else if parsed, err := time.Parse("01/02/2006", dateStr); err == nil {
				timestamp = parsed
			}
		}
		
		// Create entry
		content := strings.TrimSpace(strings.Trim(fields[contentCol], "\""))
		if content != "" {
			entry := Entry{
				ID:        generateEntryID(),
				Timestamp: timestamp,
				Content:   content,
				Type:      "imported",
			}
			
			taskMap[taskID].Entries = append(taskMap[taskID].Entries, entry)
			result.EntriesAdded++
		}
	}
	
	// Save all tasks
	for _, task := range taskMap {
		if len(task.Entries) > 0 {
			if err := js.saveTask(task); err != nil {
				warnings = append(warnings, fmt.Sprintf("Failed to save task %s: %v", task.ID, err))
			} else {
				result.TasksCreated++
			}
		}
	}
	
	result.Summary = fmt.Sprintf("Imported %d entries into %d task(s) from CSV", result.EntriesAdded, result.TasksCreated)
	return result, warnings
}

func (js *JournalService) extractTimestamp(text string) (bool, time.Time) {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-01-02 15:04",
		"01/02/2006 15:04",
	}
	
	for _, format := range formats {
		if timestamp, err := time.Parse(format, text); err == nil {
			return true, timestamp
		}
	}
	
	return false, time.Time{}
}

func (js *JournalService) parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	
	for _, char := range line {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				fields = append(fields, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}
	
	fields = append(fields, current.String())
	return fields
}

// AI-assisted recommendation analysis
func (js *JournalService) analyzeAndRecommend(tasks []*Task, taskType, focusArea string, limit int) []TaskRecommendation {
	var recommendations []TaskRecommendation
	
	// Filter tasks by type if specified
	var filteredTasks []*Task
	for _, task := range tasks {
		if taskType == "" || task.Type == taskType {
			filteredTasks = append(filteredTasks, task)
		}
	}
	
	switch focusArea {
	case "productivity":
		recommendations = js.generateProductivityRecommendations(filteredTasks, limit)
	case "learning":
		recommendations = js.generateLearningRecommendations(filteredTasks, limit)
	case "completion":
		recommendations = js.generateCompletionRecommendations(filteredTasks, limit)
	case "priority":
		recommendations = js.generatePriorityRecommendations(filteredTasks, limit)
	}
	
	return recommendations
}

func (js *JournalService) generateProductivityRecommendations(tasks []*Task, limit int) []TaskRecommendation {
	var recommendations []TaskRecommendation
	
	// Analyze task patterns
	activeTasks := js.getActiveTasks(tasks)
	staleTasks := js.getStaleTasks(tasks)
	frequentTypes := js.getFrequentTaskTypes(tasks)
	
	// Recommend breaking down large tasks
	for _, task := range activeTasks {
		if len(task.Entries) > 10 {
			recommendations = append(recommendations, TaskRecommendation{
				Type:        "task_breakdown",
				Title:       fmt.Sprintf("Break down '%s' into smaller tasks", task.Title),
				Description: "This task has many entries and might benefit from being split into smaller, more manageable tasks",
				Rationale:   fmt.Sprintf("Task has %d entries, suggesting it's complex and could be decomposed", len(task.Entries)),
				Priority:    "medium",
				Confidence:  0.7,
				SuggestedTags: []string{"breakdown", "organization"},
			})
			if len(recommendations) >= limit {
				break
			}
		}
	}
	
	// Recommend reviewing stale tasks
	for _, task := range staleTasks {
		if len(recommendations) >= limit {
			break
		}
		recommendations = append(recommendations, TaskRecommendation{
			Type:        "task_review",
			Title:       fmt.Sprintf("Review stale task: '%s'", task.Title),
			Description: "This task hasn't been updated recently and may need attention or status change",
			Rationale:   fmt.Sprintf("Last updated %s", task.Updated.Format("2006-01-02")),
			Priority:    "low",
			Confidence:  0.6,
			SuggestedTags: []string{"review", "stale"},
		})
	}
	
	// Recommend creating tasks for frequent types
	if len(recommendations) < limit && len(frequentTypes) > 0 {
		for taskType, count := range frequentTypes {
			if len(recommendations) >= limit {
				break
			}
			if count >= 3 {
				recommendations = append(recommendations, TaskRecommendation{
					Type:        "new_task",
					Title:       fmt.Sprintf("Consider creating another %s task", taskType),
					Description: fmt.Sprintf("You've been productive with %s tasks recently", taskType),
					Rationale:   fmt.Sprintf("You have %d active %s tasks showing this is an area of focus", count, taskType),
					Priority:    "medium",
					Confidence:  0.5,
					SuggestedTags: []string{taskType, "suggested"},
				})
			}
		}
	}
	
	return recommendations
}

func (js *JournalService) generateLearningRecommendations(tasks []*Task, limit int) []TaskRecommendation {
	var recommendations []TaskRecommendation
	
	learningTasks := js.getTasksByType(tasks, "learning")
	completedLearning := js.getCompletedTasks(learningTasks)
	activeLearning := js.getActiveTasks(learningTasks)
	
	// Recommend review of completed learning
	for _, task := range completedLearning {
		if len(recommendations) >= limit {
			break
		}
		if time.Since(task.Updated) > 30*24*time.Hour { // 30 days
			recommendations = append(recommendations, TaskRecommendation{
				Type:        "learning_review",
				Title:       fmt.Sprintf("Review learning: '%s'", task.Title),
				Description: "Revisiting completed learning can help reinforce knowledge",
				Rationale:   fmt.Sprintf("Completed %s, may benefit from review", task.Updated.Format("2006-01-02")),
				Priority:    "low",
				Confidence:  0.6,
				SuggestedTags: []string{"review", "learning", "reinforcement"},
			})
		}
	}
	
	// Recommend practice tasks for active learning
	for _, task := range activeLearning {
		if len(recommendations) >= limit {
			break
		}
		if len(task.Entries) >= 5 {
			recommendations = append(recommendations, TaskRecommendation{
				Type:        "practice_task",
				Title:       fmt.Sprintf("Create practice task for '%s'", task.Title),
				Description: "Apply what you've learned with a practical exercise",
				Rationale:   fmt.Sprintf("Learning task has %d entries, ready for practical application", len(task.Entries)),
				Priority:    "high",
				Confidence:  0.8,
				SuggestedTags: []string{"practice", "application", "learning"},
			})
		}
	}
	
	// Recommend new learning areas
	if len(recommendations) < limit {
		recommendations = append(recommendations, TaskRecommendation{
			Type:        "new_learning",
			Title:       "Explore a new learning area",
			Description: "Consider starting a new learning project to expand your skills",
			Rationale:   fmt.Sprintf("You have %d learning tasks, showing commitment to growth", len(learningTasks)),
			Priority:    "medium",
			Confidence:  0.4,
			SuggestedTags: []string{"new", "learning", "growth"},
		})
	}
	
	return recommendations
}

func (js *JournalService) generateCompletionRecommendations(tasks []*Task, limit int) []TaskRecommendation {
	var recommendations []TaskRecommendation
	
	activeTasks := js.getActiveTasks(tasks)
	nearCompletionTasks := js.getNearCompletionTasks(activeTasks)
	pausedTasks := js.getPausedTasks(tasks)
	
	// Recommend completing tasks that are near completion
	for _, task := range nearCompletionTasks {
		if len(recommendations) >= limit {
			break
		}
		recommendations = append(recommendations, TaskRecommendation{
			Type:        "complete_task",
			Title:       fmt.Sprintf("Complete task: '%s'", task.Title),
			Description: "This task appears to be near completion based on recent activity",
			Rationale:   fmt.Sprintf("Task has consistent recent activity (%d entries)", len(task.Entries)),
			Priority:    "high",
			Confidence:  0.8,
			SuggestedTags: []string{"completion", "finish"},
		})
	}
	
	// Recommend resuming paused tasks
	for _, task := range pausedTasks {
		if len(recommendations) >= limit {
			break
		}
		recommendations = append(recommendations, TaskRecommendation{
			Type:        "resume_task",
			Title:       fmt.Sprintf("Resume paused task: '%s'", task.Title),
			Description: "Consider resuming this paused task or updating its status",
			Rationale:   fmt.Sprintf("Task has been paused since %s", task.Updated.Format("2006-01-02")),
			Priority:    "medium",
			Confidence:  0.6,
			SuggestedTags: []string{"resume", "paused"},
		})
	}
	
	return recommendations
}

func (js *JournalService) generatePriorityRecommendations(tasks []*Task, limit int) []TaskRecommendation {
	var recommendations []TaskRecommendation
	
	urgentTasks := js.getUrgentTasks(tasks)
	oldTasks := js.getOldActiveTasks(tasks)
	
	// Recommend urgent task attention
	for _, task := range urgentTasks {
		if len(recommendations) >= limit {
			break
		}
		recommendations = append(recommendations, TaskRecommendation{
			Type:        "urgent_attention",
			Title:       fmt.Sprintf("Urgent: Focus on '%s'", task.Title),
			Description: "This task is marked as urgent and needs immediate attention",
			Rationale:   "Task has urgent priority level",
			Priority:    "urgent",
			Confidence:  0.9,
			SuggestedTags: []string{"urgent", "priority"},
		})
	}
	
	// Recommend priority review for old tasks
	for _, task := range oldTasks {
		if len(recommendations) >= limit {
			break
		}
		recommendations = append(recommendations, TaskRecommendation{
			Type:        "priority_review",
			Title:       fmt.Sprintf("Review priority of '%s'", task.Title),
			Description: "This long-running task may need priority adjustment",
			Rationale:   fmt.Sprintf("Task created %s and still active", task.Created.Format("2006-01-02")),
			Priority:    "medium",
			Confidence:  0.5,
			SuggestedTags: []string{"priority-review", "long-running"},
		})
	}
	
	return recommendations
}

// Helper functions for task analysis
func (js *JournalService) getActiveTasks(tasks []*Task) []*Task {
	var active []*Task
	for _, task := range tasks {
		if task.Status == "active" {
			active = append(active, task)
		}
	}
	return active
}

func (js *JournalService) getCompletedTasks(tasks []*Task) []*Task {
	var completed []*Task
	for _, task := range tasks {
		if task.Status == "completed" {
			completed = append(completed, task)
		}
	}
	return completed
}

func (js *JournalService) getPausedTasks(tasks []*Task) []*Task {
	var paused []*Task
	for _, task := range tasks {
		if task.Status == "paused" {
			paused = append(paused, task)
		}
	}
	return paused
}

func (js *JournalService) getTasksByType(tasks []*Task, taskType string) []*Task {
	var filtered []*Task
	for _, task := range tasks {
		if task.Type == taskType {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func (js *JournalService) getStaleTasks(tasks []*Task) []*Task {
	var stale []*Task
	cutoff := time.Now().AddDate(0, 0, -7) // 7 days ago
	for _, task := range tasks {
		if task.Status == "active" && task.Updated.Before(cutoff) {
			stale = append(stale, task)
		}
	}
	return stale
}

func (js *JournalService) getNearCompletionTasks(tasks []*Task) []*Task {
	var nearCompletion []*Task
	recent := time.Now().AddDate(0, 0, -3) // Last 3 days
	for _, task := range tasks {
		if task.Status == "active" && len(task.Entries) >= 3 && task.Updated.After(recent) {
			nearCompletion = append(nearCompletion, task)
		}
	}
	return nearCompletion
}

func (js *JournalService) getUrgentTasks(tasks []*Task) []*Task {
	var urgent []*Task
	for _, task := range tasks {
		if task.Priority == "urgent" && task.Status == "active" {
			urgent = append(urgent, task)
		}
	}
	return urgent
}

func (js *JournalService) getOldActiveTasks(tasks []*Task) []*Task {
	var old []*Task
	cutoff := time.Now().AddDate(0, 0, -30) // 30 days ago
	for _, task := range tasks {
		if task.Status == "active" && task.Created.Before(cutoff) {
			old = append(old, task)
		}
	}
	return old
}

func (js *JournalService) getFrequentTaskTypes(tasks []*Task) map[string]int {
	typeCounts := make(map[string]int)
	for _, task := range tasks {
		if task.Status == "active" {
			typeCounts[task.Type]++
		}
	}
	return typeCounts
}

func (js *JournalService) calculateAnalysisMetrics(tasks []*Task) map[string]interface{} {
	metrics := make(map[string]interface{})
	
	statusCounts := make(map[string]int)
	typeCounts := make(map[string]int)
	totalEntries := 0
	
	for _, task := range tasks {
		statusCounts[task.Status]++
		typeCounts[task.Type]++
		totalEntries += len(task.Entries)
	}
	
	metrics["total_tasks"] = len(tasks)
	metrics["total_entries"] = totalEntries
	metrics["status_breakdown"] = statusCounts
	metrics["type_breakdown"] = typeCounts
	
	if len(tasks) > 0 {
		metrics["avg_entries_per_task"] = float64(totalEntries) / float64(len(tasks))
	}
	
	return metrics
}

// Analytics helper functions
func (js *JournalService) filterTasksByTimePeriod(tasks []*Task, timePeriod string) []*Task {
	if timePeriod == "all" {
		return tasks
	}
	
	var cutoffDate time.Time
	now := time.Now()
	
	switch timePeriod {
	case "week":
		cutoffDate = now.AddDate(0, 0, -7)
	case "month":
		cutoffDate = now.AddDate(0, -1, 0)
	case "quarter":
		cutoffDate = now.AddDate(0, -3, 0)
	case "year":
		cutoffDate = now.AddDate(-1, 0, 0)
	default:
		return tasks
	}
	
	var filtered []*Task
	for _, task := range tasks {
		if task.Updated.After(cutoffDate) || task.Created.After(cutoffDate) {
			filtered = append(filtered, task)
		}
	}
	
	return filtered
}

func (js *JournalService) generateAnalyticsReport(tasks []*Task, reportType, timePeriod string) AnalyticsReport {
	report := AnalyticsReport{
		ReportType:  reportType,
		TimePeriod:  timePeriod,
		GeneratedAt: time.Now(),
		TaskMetrics: js.calculateTaskMetrics(tasks),
		ProductivityMetrics: js.calculateProductivityMetrics(tasks, timePeriod),
		PatternAnalysis: js.calculatePatternAnalysis(tasks),
		Insights: js.generateInsights(tasks, reportType),
	}
	
	if reportType == "trends" || reportType == "overview" {
		report.Trends = js.calculateTrends(tasks, timePeriod)
	}
	
	report.Summary = js.generateReportSummary(report, tasks)
	
	return report
}

func (js *JournalService) calculateTaskMetrics(tasks []*Task) TaskMetrics {
	metrics := TaskMetrics{
		TotalTasks: len(tasks),
		ByStatus:   make(map[string]int),
		ByType:     make(map[string]int),
		ByPriority: make(map[string]int),
	}
	
	totalEntries := 0
	completedTasks := 0
	
	for _, task := range tasks {
		metrics.ByStatus[task.Status]++
		metrics.ByType[task.Type]++
		
		priority := task.Priority
		if priority == "" {
			priority = "none"
		}
		metrics.ByPriority[priority]++
		
		totalEntries += len(task.Entries)
		
		if task.Status == "completed" {
			completedTasks++
		}
	}
	
	metrics.TotalEntries = totalEntries
	
	if len(tasks) > 0 {
		metrics.AverageEntries = float64(totalEntries) / float64(len(tasks))
		metrics.CompletionRate = float64(completedTasks) / float64(len(tasks))
	}
	
	return metrics
}

func (js *JournalService) calculateProductivityMetrics(tasks []*Task, timePeriod string) ProductivityMetrics {
	metrics := ProductivityMetrics{}
	
	var completedTasks []*Task
	var totalDuration float64
	var durationCount int
	typeEntries := make(map[string]int)
	entriesInPeriod := 0
	
	// Calculate time period bounds
	now := time.Now()
	var periodStart time.Time
	switch timePeriod {
	case "week":
		periodStart = now.AddDate(0, 0, -7)
	case "month":
		periodStart = now.AddDate(0, -1, 0)
	case "quarter":
		periodStart = now.AddDate(0, -3, 0)
	case "year":
		periodStart = now.AddDate(-1, 0, 0)
	default:
		periodStart = time.Time{} // All time
	}
	
	for _, task := range tasks {
		// Count entries in period
		for _, entry := range task.Entries {
			if timePeriod == "all" || entry.Timestamp.After(periodStart) {
				entriesInPeriod++
				typeEntries[task.Type]++
			}
		}
		
		// Track completed tasks
		if task.Status == "completed" {
			completedTasks = append(completedTasks, task)
			if timePeriod == "all" || task.Updated.After(periodStart) {
				metrics.TasksCompletedPeriod++
			}
			
			// Calculate task duration
			duration := task.Updated.Sub(task.Created).Hours() / 24 // Convert to days
			if duration > 0 {
				totalDuration += duration
				durationCount++
			}
		}
	}
	
	metrics.EntriesAddedPeriod = entriesInPeriod
	
	if durationCount > 0 {
		metrics.AverageTaskDuration = totalDuration / float64(durationCount)
	}
	
	// Find most productive type
	maxEntries := 0
	for taskType, entries := range typeEntries {
		if entries > maxEntries {
			maxEntries = entries
			metrics.MostProductiveType = taskType
		}
	}
	
	// Calculate productivity score (arbitrary metric)
	if len(tasks) > 0 {
		metrics.ProductivityScore = (float64(metrics.TasksCompletedPeriod) * 2 + 
			float64(metrics.EntriesAddedPeriod) * 0.1) / float64(len(tasks))
	}
	
	return metrics
}

func (js *JournalService) calculatePatternAnalysis(tasks []*Task) PatternAnalysis {
	analysis := PatternAnalysis{
		WorkPatterns:      make(map[string]int),
		TimeToCompletion:  make(map[string]float64),
	}
	
	typeCounts := make(map[string]int)
	tagCounts := make(map[string]int)
	typeCompletionTimes := make(map[string][]float64)
	
	for _, task := range tasks {
		typeCounts[task.Type]++
		
		// Count tags
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}
		
		// Analyze work patterns (entry frequency)
		if len(task.Entries) > 0 {
			// Simple pattern: tasks with many entries vs few entries
			if len(task.Entries) >= 5 {
				analysis.WorkPatterns["intensive"]++
			} else {
				analysis.WorkPatterns["light"]++
			}
		}
		
		// Calculate completion time by type
		if task.Status == "completed" {
			duration := task.Updated.Sub(task.Created).Hours() / 24 // Days
			if duration > 0 {
				typeCompletionTimes[task.Type] = append(typeCompletionTimes[task.Type], duration)
			}
		}
	}
	
	// Find most frequent type
	maxCount := 0
	for taskType, count := range typeCounts {
		if count > maxCount {
			maxCount = count
			analysis.MostFrequentType = taskType
		}
	}
	
	// Get common tags (more than 1 occurrence)
	for tag, count := range tagCounts {
		if count > 1 {
			analysis.CommonTags = append(analysis.CommonTags, tag)
		}
	}
	
	// Calculate average completion time by type
	for taskType, durations := range typeCompletionTimes {
		if len(durations) > 0 {
			sum := 0.0
			for _, duration := range durations {
				sum += duration
			}
			analysis.TimeToCompletion[taskType] = sum / float64(len(durations))
		}
	}
	
	return analysis
}

func (js *JournalService) calculateTrends(tasks []*Task, timePeriod string) []Trend {
	var trends []Trend
	
	// Simple trend calculation comparing recent vs previous period
	now := time.Now()
	var recentStart, previousStart time.Time
	
	switch timePeriod {
	case "week":
		recentStart = now.AddDate(0, 0, -7)
		previousStart = now.AddDate(0, 0, -14)
	case "month":
		recentStart = now.AddDate(0, -1, 0)
		previousStart = now.AddDate(0, -2, 0)
	case "quarter":
		recentStart = now.AddDate(0, -3, 0)
		previousStart = now.AddDate(0, -6, 0)
	case "year":
		recentStart = now.AddDate(-1, 0, 0)
		previousStart = now.AddDate(-2, 0, 0)
	default:
		return trends // No trends for "all" period
	}
	
	recentTasks := 0
	previousTasks := 0
	recentEntries := 0
	previousEntries := 0
	
	for _, task := range tasks {
		// Count tasks created in periods
		if task.Created.After(recentStart) {
			recentTasks++
		} else if task.Created.After(previousStart) && task.Created.Before(recentStart) {
			previousTasks++
		}
		
		// Count entries in periods
		for _, entry := range task.Entries {
			if entry.Timestamp.After(recentStart) {
				recentEntries++
			} else if entry.Timestamp.After(previousStart) && entry.Timestamp.Before(recentStart) {
				previousEntries++
			}
		}
	}
	
	// Calculate task creation trend
	if previousTasks > 0 {
		change := float64(recentTasks-previousTasks) / float64(previousTasks) * 100
		direction := "stable"
		if change > 5 {
			direction = "up"
		} else if change < -5 {
			direction = "down"
		}
		
		trends = append(trends, Trend{
			Metric:    "task_creation",
			Direction: direction,
			Change:    change,
			Period:    timePeriod,
		})
	}
	
	// Calculate activity trend
	if previousEntries > 0 {
		change := float64(recentEntries-previousEntries) / float64(previousEntries) * 100
		direction := "stable"
		if change > 10 {
			direction = "up"
		} else if change < -10 {
			direction = "down"
		}
		
		trends = append(trends, Trend{
			Metric:    "activity_level",
			Direction: direction,
			Change:    change,
			Period:    timePeriod,
		})
	}
	
	return trends
}

func (js *JournalService) generateInsights(tasks []*Task, reportType string) []string {
	var insights []string
	
	if len(tasks) == 0 {
		return []string{"No tasks available for analysis"}
	}
	
	// General insights
	activeTasks := js.getActiveTasks(tasks)
	completedTasks := js.getCompletedTasks(tasks)
	
	if len(activeTasks) > len(completedTasks)*2 {
		insights = append(insights, "You have many active tasks relative to completed ones. Consider focusing on completion.")
	}
	
	// Type-based insights
	typeCounts := make(map[string]int)
	for _, task := range tasks {
		typeCounts[task.Type]++
	}
	
	maxType := ""
	maxCount := 0
	for taskType, count := range typeCounts {
		if count > maxCount {
			maxCount = count
			maxType = taskType
		}
	}
	
	if maxType != "" {
		insights = append(insights, fmt.Sprintf("Your primary focus area is '%s' tasks (%d tasks)", maxType, maxCount))
	}
	
	// Productivity insights
	staleTasks := js.getStaleTasks(tasks)
	if len(staleTasks) > 0 {
		insights = append(insights, fmt.Sprintf("You have %d stale tasks that haven't been updated recently", len(staleTasks)))
	}
	
	// Entry patterns
	totalEntries := 0
	for _, task := range tasks {
		totalEntries += len(task.Entries)
	}
	
	if totalEntries > 0 && len(tasks) > 0 {
		avgEntries := float64(totalEntries) / float64(len(tasks))
		if avgEntries > 10 {
			insights = append(insights, "Your tasks tend to have many entries, suggesting detailed tracking")
		} else if avgEntries < 3 {
			insights = append(insights, "Your tasks have few entries on average, consider more detailed logging")
		}
	}
	
	return insights
}

func (js *JournalService) generateReportSummary(report AnalyticsReport, tasks []*Task) string {
	summary := fmt.Sprintf("Analytics report for %s period: ", report.TimePeriod)
	summary += fmt.Sprintf("%d total tasks, ", report.TaskMetrics.TotalTasks)
	summary += fmt.Sprintf("%.1f%% completion rate, ", report.TaskMetrics.CompletionRate*100)
	summary += fmt.Sprintf("%d total entries. ", report.TaskMetrics.TotalEntries)
	
	if report.ProductivityMetrics.MostProductiveType != "" {
		summary += fmt.Sprintf("Most active area: %s. ", report.ProductivityMetrics.MostProductiveType)
	}
	
	if len(report.Trends) > 0 {
		summary += fmt.Sprintf("Detected %d trends in your work patterns.", len(report.Trends))
	}
	
	return summary
}
