package main

import (
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer("journal-mcp", "1.0.0", server.WithToolCapabilities(true))

	// Initialize the journal service
	journalService := NewJournalService()

	// Register tools
	registerTools(s, journalService)

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}

func registerTools(s *server.MCPServer, js *JournalService) {
	// Task Management Tools
	s.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task with optional issue linking"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Unique task identifier (e.g., MDU-1450, learning-graphql)"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Task title or description"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Task type: work, learning, personal, investigation"),
		),
		mcp.WithArray("tags",
			mcp.Description("Flat tags for categorization"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithString("issue_url",
			mcp.Description("Full URL to GitHub issue or Jira ticket"),
		),
		mcp.WithString("priority",
			mcp.Description("Priority level: low, medium, high, urgent"),
		),
	), js.CreateTask)

	s.AddTool(mcp.NewTool("add_task_entry",
		mcp.WithDescription("Add a timestamped entry to an existing task"),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("Task identifier"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Entry content"),
		),
		mcp.WithString("timestamp",
			mcp.Description("ISO timestamp (defaults to now)"),
		),
	), js.AddTaskEntry)

	s.AddTool(mcp.NewTool("get_task",
		mcp.WithDescription("Retrieve complete task history"),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("Task identifier"),
		),
	), js.GetTask)

	s.AddTool(mcp.NewTool("list_tasks",
		mcp.WithDescription("List tasks with optional filtering and pagination"),
		mcp.WithString("status",
			mcp.Description("Filter by status: active, completed, paused"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by type: work, learning, personal, investigation"),
		),
		mcp.WithString("date_from",
			mcp.Description("Filter tasks updated from this date (YYYY-MM-DD format)"),
		),
		mcp.WithString("date_to",
			mcp.Description("Filter tasks updated until this date (YYYY-MM-DD format)"),
		),
		mcp.WithString("limit",
			mcp.Description("Maximum number of tasks to return (default: 50, max: 200)"),
		),
		mcp.WithString("offset",
			mcp.Description("Number of tasks to skip for pagination (default: 0)"),
		),
	), js.ListTasks)

	s.AddTool(mcp.NewTool("update_task_status",
		mcp.WithDescription("Change task status (active/completed/paused/blocked)"),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("Task identifier"),
		),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("New status: active, completed, paused, blocked"),
		),
		mcp.WithString("reason",
			mcp.Description("Optional reason for status change"),
		),
	), js.UpdateTaskStatus)

	// Daily and Weekly Logs
	s.AddTool(mcp.NewTool("get_daily_log",
		mcp.WithDescription("View all activity for a specific date"),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("Date in YYYY-MM-DD format"),
		),
	), js.GetDailyLog)

	s.AddTool(mcp.NewTool("get_weekly_log",
		mcp.WithDescription("View activity for a week (aggregate daily logs)"),
		mcp.WithString("week_start",
			mcp.Required(),
			mcp.Description("Week start date in YYYY-MM-DD format"),
		),
	), js.GetWeeklyLog)

	// One-on-One Meeting Tools
	s.AddTool(mcp.NewTool("create_one_on_one",
		mcp.WithDescription("Record structured meeting notes"),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("Meeting date in YYYY-MM-DD format"),
		),
		mcp.WithArray("insights",
			mcp.Description("Key insights from the meeting"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithArray("todos",
			mcp.Description("Action items and todos"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithArray("feedback",
			mcp.Description("Feedback points"),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithString("notes",
			mcp.Description("Additional meeting notes"),
		),
	), js.CreateOneOnOne)

	s.AddTool(mcp.NewTool("get_one_on_one_history",
		mcp.WithDescription("Retrieve meeting history"),
		mcp.WithString("limit",
			mcp.Description("Number of meetings to retrieve (default: 10)"),
		),
	), js.GetOneOnOneHistory)

	// Search and Export Tools
	s.AddTool(mcp.NewTool("search_entries",
		mcp.WithDescription("Search through all journal content"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query text"),
		),
		mcp.WithString("task_type",
			mcp.Description("Filter by task type: work, learning, personal, investigation"),
		),
		mcp.WithString("date_from",
			mcp.Description("Start date filter (YYYY-MM-DD)"),
		),
		mcp.WithString("date_to",
			mcp.Description("End date filter (YYYY-MM-DD)"),
		),
	), js.SearchEntries)

	s.AddTool(mcp.NewTool("export_data",
		mcp.WithDescription("Export journal data to various formats"),
		mcp.WithString("format",
			mcp.Required(),
			mcp.Description("Export format: json, markdown, csv"),
		),
		mcp.WithString("date_from",
			mcp.Description("Start date filter (YYYY-MM-DD)"),
		),
		mcp.WithString("date_to",
			mcp.Description("End date filter (YYYY-MM-DD)"),
		),
		mcp.WithString("task_filter",
			mcp.Description("Filter by task type: work, learning, personal, investigation"),
		),
	), js.ExportData)
}
