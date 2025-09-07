package server

import (
	"github.com/cpuchip/journal-mcp/internal/journal"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewMCPServer creates and configures a new MCP server with all tools registered
func NewMCPServer(js *journal.Service) *server.MCPServer {
	s := server.NewMCPServer("journal-mcp", "1.0.0", server.WithToolCapabilities(true))
	registerTools(s, js)
	return s
}

func registerTools(s *server.MCPServer, js *journal.Service) {
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

	// Import and Analytics Tools
	s.AddTool(mcp.NewTool("import_data",
		mcp.WithDescription("Import existing diary/journal data from various formats"),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("File content to import"),
		),
		mcp.WithString("format",
			mcp.Required(),
			mcp.Description("Input format: txt, markdown, json, csv"),
		),
		mcp.WithString("task_prefix",
			mcp.Description("Optional prefix for auto-generated task IDs (default: 'IMPORT')"),
		),
		mcp.WithString("default_type",
			mcp.Description("Default task type for imported entries: work, learning, personal, investigation (default: 'personal')"),
		),
	), js.ImportData)

	s.AddTool(mcp.NewTool("get_task_recommendations",
		mcp.WithDescription("Get AI-assisted task recommendations based on patterns and history"),
		mcp.WithString("task_type",
			mcp.Description("Filter recommendations by task type: work, learning, personal, investigation"),
		),
		mcp.WithString("focus_area",
			mcp.Description("Focus area for recommendations: productivity, learning, completion, priority"),
		),
		mcp.WithString("limit",
			mcp.Description("Maximum number of recommendations to return (default: 5, max: 20)"),
		),
	), js.GetTaskRecommendations)

	s.AddTool(mcp.NewTool("get_analytics_report",
		mcp.WithDescription("Generate comprehensive analytics and insights report"),
		mcp.WithString("report_type",
			mcp.Description("Type of report: overview, productivity, patterns, trends (default: overview)"),
		),
		mcp.WithString("time_period",
			mcp.Description("Time period for analysis: week, month, quarter, year, all (default: month)"),
		),
		mcp.WithString("task_type",
			mcp.Description("Filter by task type: work, learning, personal, investigation"),
		),
	), js.GetAnalyticsReport)

	// GitHub Integration Tools
	s.AddTool(mcp.NewTool("sync_with_github",
		mcp.WithDescription("Sync assigned GitHub issues with tasks"),
		mcp.WithString("github_token",
			mcp.Required(),
			mcp.Description("GitHub personal access token"),
		),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("GitHub username to sync issues for"),
		),
		mcp.WithArray("repositories",
			mcp.Description("Optional list of repositories to sync (format: owner/repo). If empty, syncs all assigned issues."),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithString("create_tasks",
			mcp.Description("Whether to create new tasks for new issues (true/false, default: true)"),
		),
		mcp.WithString("update_existing",
			mcp.Description("Whether to update existing tasks with issue changes (true/false, default: true)"),
		),
	), js.SyncWithGitHub)

	s.AddTool(mcp.NewTool("pull_issue_updates",
		mcp.WithDescription("Pull latest comments and events for tracked GitHub issues"),
		mcp.WithString("github_token",
			mcp.Required(),
			mcp.Description("GitHub personal access token"),
		),
		mcp.WithString("task_id",
			mcp.Description("Specific task ID to update (if empty, updates all tasks with GitHub issues)"),
		),
		mcp.WithString("since",
			mcp.Description("Only pull updates since this timestamp (ISO 8601 format: 2006-01-02T15:04:05Z)"),
		),
	), js.PullIssueUpdates)

	s.AddTool(mcp.NewTool("create_task_from_github_issue",
		mcp.WithDescription("Create a new task from a GitHub issue URL"),
		mcp.WithString("github_token",
			mcp.Required(),
			mcp.Description("GitHub personal access token"),
		),
		mcp.WithString("issue_url",
			mcp.Required(),
			mcp.Description("Full GitHub issue URL (e.g., https://github.com/owner/repo/issues/123)"),
		),
		mcp.WithString("type",
			mcp.Description("Task type: work, learning, personal, investigation (default: work)"),
		),
		mcp.WithString("priority",
			mcp.Description("Task priority: low, medium, high, urgent (default: medium)"),
		),
	), js.CreateTaskFromGitHubIssue)

	// Data Management Tools
	s.AddTool(mcp.NewTool("create_data_backup",
		mcp.WithDescription("Create a backup of all journal data"),
		mcp.WithString("backup_path",
			mcp.Description("Path for the backup file (defaults to timestamped file in data directory)"),
		),
		mcp.WithString("include_config",
			mcp.Description("Whether to include configuration in backup (true/false, default: true)"),
		),
		mcp.WithString("compression",
			mcp.Description("Compression level: none, default, maximum (default: default)"),
		),
	), js.CreateDataBackup)

	s.AddTool(mcp.NewTool("restore_data_backup",
		mcp.WithDescription("Restore journal data from a backup file"),
		mcp.WithString("backup_path",
			mcp.Required(),
			mcp.Description("Path to the backup ZIP file"),
		),
		mcp.WithString("overwrite_existing",
			mcp.Description("Whether to overwrite existing data (true/false, default: false)"),
		),
		mcp.WithString("restore_config",
			mcp.Description("Whether to restore configuration (true/false, default: true)"),
		),
	), js.RestoreDataBackup)

	s.AddTool(mcp.NewTool("get_configuration",
		mcp.WithDescription("Get the current journal configuration"),
	), js.GetConfiguration)

	s.AddTool(mcp.NewTool("update_configuration",
		mcp.WithDescription("Update the journal configuration"),
		mcp.WithString("config",
			mcp.Required(),
			mcp.Description("Configuration JSON data"),
		),
	), js.UpdateConfiguration)

	s.AddTool(mcp.NewTool("migrate_data",
		mcp.WithDescription("Perform data migration (future SQLite integration preparation)"),
		mcp.WithString("target_version",
			mcp.Description("Target migration version (default: current)"),
		),
		mcp.WithString("dry_run",
			mcp.Description("Perform a dry run without making changes (true/false, default: false)"),
		),
	), js.MigrateData)
}