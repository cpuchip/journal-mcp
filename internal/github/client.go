package github

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/oauth2"
)

// GitHubService handles GitHub API interactions
type GitHubService struct {
	client *github.Client
	token  string
}

// GitHubSyncConfig holds configuration for GitHub sync
type GitHubSyncConfig struct {
	Token        string   `json:"token" yaml:"token"`
	Username     string   `json:"username" yaml:"username"`
	Repositories []string `json:"repositories" yaml:"repositories"`
	AutoSync     bool     `json:"auto_sync" yaml:"auto_sync"`
	SyncInterval int      `json:"sync_interval_minutes" yaml:"sync_interval_minutes"` // in minutes
}

// GitHubIssueMetadata stores additional GitHub-specific data
type GitHubIssueMetadata struct {
	IssueNumber int                        `json:"issue_number"`
	Repository  string                     `json:"repository"`
	State       string                     `json:"state"`
	Labels      []string                   `json:"labels"`
	Assignees   []string                   `json:"assignees"`
	Milestone   string                     `json:"milestone,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	Comments    []GitHubIssueComment       `json:"comments,omitempty"`
	Events      []GitHubIssueEvent         `json:"events,omitempty"`
}

// GitHubIssueComment represents a comment on a GitHub issue
type GitHubIssueComment struct {
	ID        int64     `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GitHubIssueEvent represents an event on a GitHub issue
type GitHubIssueEvent struct {
	ID        int64     `json:"id"`
	Event     string    `json:"event"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
	Label     string    `json:"label,omitempty"`
}

// GitHubSyncResult represents the result of a GitHub sync operation
type GitHubSyncResult struct {
	TasksCreated    int      `json:"tasks_created"`
	TasksUpdated    int      `json:"tasks_updated"`
	IssuesProcessed int      `json:"issues_processed"`
	Errors          []string `json:"errors,omitempty"`
	Summary         string   `json:"summary"`
	LastSyncTime    time.Time `json:"last_sync_time"`
}

// NewGitHubService creates a new GitHub service with authentication
func NewGitHubService(token string) *GitHubService {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	
	return &GitHubService{
		client: github.NewClient(tc),
		token:  token,
	}
}

// SyncWithGitHub syncs assigned GitHub issues with tasks
func (js *JournalService) SyncWithGitHub(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token := request.GetString("github_token", "")
	if token == "" {
		return mcp.NewToolResultError("github_token is required"), nil
	}

	username := request.GetString("username", "")
	if username == "" {
		return mcp.NewToolResultError("username is required"), nil
	}

	repositories := request.GetStringSlice("repositories", nil)
	createTasks := request.GetString("create_tasks", "true") == "true"
	updateExisting := request.GetString("update_existing", "true") == "true"

	githubService := NewGitHubService(token)
	
	syncResult := GitHubSyncResult{
		LastSyncTime: time.Now(),
		Errors:       []string{},
	}

	// Get assigned issues from GitHub
	issues, err := githubService.getAssignedIssues(ctx, username, repositories)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch GitHub issues: %v", err)), nil
	}

	syncResult.IssuesProcessed = len(issues)

	for _, issue := range issues {
		taskID := generateTaskIDFromIssue(issue)
		
		// Check if task already exists
		existingTask, err := js.loadTask(taskID)
		if err != nil {
			// Task doesn't exist, create new one if enabled
			if createTasks {
				task := js.createTaskFromGitHubIssue(issue)
				if err := js.saveTask(task); err != nil {
					syncResult.Errors = append(syncResult.Errors, fmt.Sprintf("Failed to save task %s: %v", taskID, err))
					continue
				}
				syncResult.TasksCreated++
			}
		} else {
			// Task exists, update if enabled
			if updateExisting {
				updated := js.updateTaskFromGitHubIssue(existingTask, issue)
				if updated {
					if err := js.saveTask(existingTask); err != nil {
						syncResult.Errors = append(syncResult.Errors, fmt.Sprintf("Failed to update task %s: %v", taskID, err))
						continue
					}
					syncResult.TasksUpdated++
				}
			}
		}
	}

	syncResult.Summary = fmt.Sprintf("Processed %d issues: %d tasks created, %d tasks updated", 
		syncResult.IssuesProcessed, syncResult.TasksCreated, syncResult.TasksUpdated)

	result, _ := json.Marshal(syncResult)
	return mcp.NewToolResultText(string(result)), nil
}

// PullIssueUpdates pulls latest comments and events for tracked GitHub issues
func (js *JournalService) PullIssueUpdates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token := request.GetString("github_token", "")
	if token == "" {
		return mcp.NewToolResultError("github_token is required"), nil
	}

	taskID := request.GetString("task_id", "")
	sinceStr := request.GetString("since", "")
	
	var since *time.Time
	if sinceStr != "" {
		sinceTime, err := time.Parse("2006-01-02T15:04:05Z", sinceStr)
		if err != nil {
			return mcp.NewToolResultError("Invalid since timestamp format. Use ISO 8601 (2006-01-02T15:04:05Z)"), nil
		}
		since = &sinceTime
	}

	githubService := NewGitHubService(token)
	
	// Get all tasks with GitHub issues or specific task
	var tasks []*Task
	if taskID != "" {
		task, err := js.loadTask(taskID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Task not found: %s", taskID)), nil
		}
		if task.IssueURL == "" {
			return mcp.NewToolResultError("Task does not have a GitHub issue associated"), nil
		}
		tasks = []*Task{task}
	} else {
		allTasks, err := js.getAllTasks()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load tasks: %v", err)), nil
		}
		// Filter tasks that have GitHub issue URLs
		for _, task := range allTasks {
			if task.IssueURL != "" && strings.Contains(task.IssueURL, "github.com") {
				tasks = append(tasks, task)
			}
		}
	}

	updateCount := 0
	errors := []string{}

	for _, task := range tasks {
		owner, repo, issueNum, err := parseGitHubURL(task.IssueURL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid GitHub URL for task %s: %v", task.ID, err))
			continue
		}

		// Get issue comments and events
		comments, events, err := githubService.getIssueUpdates(ctx, owner, repo, issueNum, since)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to fetch updates for task %s: %v", task.ID, err))
			continue
		}

		// Add new entries for comments and events
		entriesAdded := 0
		for _, comment := range comments {
			entry := Entry{
				ID:        generateEntryID(),
				Timestamp: comment.CreatedAt,
				Content:   fmt.Sprintf("GitHub comment by %s: %s", comment.Author, comment.Body),
				Type:      "github_comment",
			}
			task.Entries = append(task.Entries, entry)
			entriesAdded++
		}

		for _, event := range events {
			entry := Entry{
				ID:        generateEntryID(),
				Timestamp: event.CreatedAt,
				Content:   fmt.Sprintf("GitHub event: %s by %s", event.Event, event.Actor),
				Type:      "github_event",
			}
			task.Entries = append(task.Entries, entry)
			entriesAdded++
		}

		if entriesAdded > 0 {
			task.Updated = time.Now()
			if err := js.saveTask(task); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to save task %s: %v", task.ID, err))
				continue
			}
			updateCount++
		}
	}

	result := map[string]interface{}{
		"tasks_updated":   updateCount,
		"total_tasks":     len(tasks),
		"errors":          errors,
		"summary":         fmt.Sprintf("Updated %d tasks with latest GitHub activity", updateCount),
		"last_sync_time":  time.Now(),
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// CreateTaskFromGitHubIssue creates a new task from a GitHub issue URL
func (js *JournalService) CreateTaskFromGitHubIssue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token := request.GetString("github_token", "")
	if token == "" {
		return mcp.NewToolResultError("github_token is required"), nil
	}

	issueURL := request.GetString("issue_url", "")
	if issueURL == "" {
		return mcp.NewToolResultError("issue_url is required"), nil
	}

	taskType := request.GetString("type", "work")
	priority := request.GetString("priority", "medium")

	githubService := NewGitHubService(token)
	
	owner, repo, issueNum, err := parseGitHubURL(issueURL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid GitHub URL: %v", err)), nil
	}

	// Fetch issue details from GitHub
	issue, err := githubService.getIssue(ctx, owner, repo, issueNum)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to fetch GitHub issue: %v", err)), nil
	}

	task := js.createTaskFromGitHubIssue(issue)
	task.Type = taskType
	task.Priority = priority

	if err := js.saveTask(task); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save task: %v", err)), nil
	}

	result := map[string]interface{}{
		"task_id":     task.ID,
		"title":       task.Title,
		"status":      task.Status,
		"issue_url":   task.IssueURL,
		"created_at":  task.Created,
		"summary":     fmt.Sprintf("Created task %s from GitHub issue %s", task.ID, issueURL),
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// Helper methods for GitHub service

func (gs *GitHubService) getAssignedIssues(ctx context.Context, username string, repositories []string) ([]*github.Issue, error) {
	var allIssues []*github.Issue

	if len(repositories) == 0 {
		// Search across all repositories assigned to user
		query := fmt.Sprintf("assignee:%s state:open", username)
		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}

		for {
			result, resp, err := gs.client.Search.Issues(ctx, query, opts)
			if err != nil {
				return nil, err
			}

			allIssues = append(allIssues, result.Issues...)

			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
	} else {
		// Search in specific repositories
		for _, repo := range repositories {
			parts := strings.Split(repo, "/")
			if len(parts) != 2 {
				continue
			}

			opts := &github.IssueListByRepoOptions{
				Assignee:    username,
				State:       "open",
				ListOptions: github.ListOptions{PerPage: 100},
			}

			for {
				issues, resp, err := gs.client.Issues.ListByRepo(ctx, parts[0], parts[1], opts)
				if err != nil {
					return nil, err
				}

				allIssues = append(allIssues, issues...)

				if resp.NextPage == 0 {
					break
				}
				opts.Page = resp.NextPage
			}
		}
	}

	return allIssues, nil
}

func (gs *GitHubService) getIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issue, _, err := gs.client.Issues.Get(ctx, owner, repo, number)
	return issue, err
}

func (gs *GitHubService) getIssueUpdates(ctx context.Context, owner, repo string, number int, since *time.Time) ([]GitHubIssueComment, []GitHubIssueEvent, error) {
	var comments []GitHubIssueComment
	var events []GitHubIssueEvent

	// Get comments
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if since != nil {
		opts.Since = since
	}

	for {
		githubComments, resp, err := gs.client.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, nil, err
		}

		for _, comment := range githubComments {
			comments = append(comments, GitHubIssueComment{
				ID:        comment.GetID(),
				Author:    comment.GetUser().GetLogin(),
				Body:      comment.GetBody(),
				CreatedAt: comment.GetCreatedAt().Time,
				UpdatedAt: comment.GetUpdatedAt().Time,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Get events
	eventOpts := &github.ListOptions{PerPage: 100}
	for {
		githubEvents, resp, err := gs.client.Issues.ListIssueEvents(ctx, owner, repo, number, eventOpts)
		if err != nil {
			return comments, events, nil // Return comments even if events fail
		}

		for _, event := range githubEvents {
			if since != nil && event.GetCreatedAt().Time.Before(*since) {
				continue
			}

			eventItem := GitHubIssueEvent{
				ID:        event.GetID(),
				Event:     event.GetEvent(),
				Actor:     event.GetActor().GetLogin(),
				CreatedAt: event.GetCreatedAt().Time,
			}

			if event.GetLabel() != nil {
				eventItem.Label = event.GetLabel().GetName()
			}

			events = append(events, eventItem)
		}

		if resp.NextPage == 0 {
			break
		}
		eventOpts.Page = resp.NextPage
	}

	return comments, events, nil
}

func (js *JournalService) createTaskFromGitHubIssue(issue *github.Issue) *Task {
	taskID := generateTaskIDFromIssue(issue)
	
	// Extract labels
	var labels []string
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}

	// Determine priority from labels
	priority := "medium"
	for _, label := range labels {
		switch strings.ToLower(label) {
		case "priority/high", "high", "critical":
			priority = "high"
		case "priority/low", "low":
			priority = "low"
		case "priority/urgent", "urgent":
			priority = "urgent"
		}
	}

	task := &Task{
		ID:       taskID,
		Title:    issue.GetTitle(),
		Type:     "work", // Default, can be overridden
		Tags:     labels,
		Status:   mapGitHubStateToTaskStatus(issue.GetState()),
		Priority: priority,
		IssueURL: issue.GetHTMLURL(),
		IssueID:  strconv.Itoa(issue.GetNumber()),
		Created:  time.Now(),
		Updated:  time.Now(),
		Entries:  []Entry{},
	}

	// Add initial entry with issue description
	if issue.GetBody() != "" {
		task.Entries = append(task.Entries, Entry{
			ID:        generateEntryID(),
			Timestamp: time.Now(),
			Content:   fmt.Sprintf("GitHub Issue Description: %s", issue.GetBody()),
			Type:      "github_description",
		})
	}

	return task
}

func (js *JournalService) updateTaskFromGitHubIssue(task *Task, issue *github.Issue) bool {
	updated := false

	// Update status if changed
	newStatus := mapGitHubStateToTaskStatus(issue.GetState())
	if task.Status != newStatus {
		task.Status = newStatus
		task.Updated = time.Now()
		
		// Add status change entry
		task.Entries = append(task.Entries, Entry{
			ID:        generateEntryID(),
			Timestamp: time.Now(),
			Content:   fmt.Sprintf("Status updated from GitHub: %s", newStatus),
			Type:      "status_change",
		})
		updated = true
	}

	// Update title if changed
	if task.Title != issue.GetTitle() {
		task.Title = issue.GetTitle()
		task.Updated = time.Now()
		updated = true
	}

	// Update labels/tags
	var newLabels []string
	for _, label := range issue.Labels {
		newLabels = append(newLabels, label.GetName())
	}
	
	if !equalStringSlices(task.Tags, newLabels) {
		task.Tags = newLabels
		task.Updated = time.Now()
		updated = true
	}

	return updated
}

// Helper functions

func generateTaskIDFromIssue(issue *github.Issue) string {
	repo := issue.GetRepository()
	if repo != nil {
		return fmt.Sprintf("GH-%s-%d", repo.GetName(), issue.GetNumber())
	}
	return fmt.Sprintf("GH-issue-%d", issue.GetNumber())
}

func parseGitHubURL(url string) (owner, repo string, number int, err error) {
	// Parse URLs like: https://github.com/owner/repo/issues/123
	parts := strings.Split(url, "/")
	if len(parts) < 7 || !strings.Contains(url, "github.com") {
		return "", "", 0, fmt.Errorf("invalid GitHub URL format")
	}

	owner = parts[3]
	repo = parts[4]
	number, err = strconv.Atoi(parts[6])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid issue number")
	}

	return owner, repo, number, nil
}

func mapGitHubStateToTaskStatus(state string) string {
	switch state {
	case "open":
		return "active"
	case "closed":
		return "completed"
	default:
		return "active"
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	sortedA := make([]string, len(a))
	sortedB := make([]string, len(b))
	copy(sortedA, a)
	copy(sortedB, b)
	
	sort.Strings(sortedA)
	sort.Strings(sortedB)
	
	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}
	
	return true
}