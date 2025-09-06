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
		mcp.WithDescription("List tasks with optional filtering"),
		mcp.WithString("status",
			mcp.Description("Filter by status: active, completed, paused"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by type: work, learning, personal, investigation"),
		),
	), js.ListTasks)
}
