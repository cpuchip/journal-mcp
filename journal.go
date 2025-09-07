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
