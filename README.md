# Journal MCP

A Model Context Protocol (MCP) server for task-based work journaling with AI integration.

## Features

- **Task-based logging** - Organize work by tasks/issues rather than just chronologically
- **Issue integration** - Link tasks to GitHub issues or Jira tickets  
- **Structured 1-on-1s** - Capture meeting insights, todos, and feedback
- **Time-based views** - Daily and weekly activity summaries
- **Tagging system** - Flat tag structure for easy categorization
- **Human-readable storage** - All data stored as JSON and Markdown
- **Search and export** - Find entries and export to various formats
- **Web interface** - View and edit journal entries through a web UI

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/cpuchip/journal-mcp.git
   cd journal-mcp
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the server:
   ```bash
   go build -o journal-mcp
   ```

## Usage

### As MCP Server

Run the server in stdio mode for MCP integration:
```bash
./journal-mcp
```

### Configuration

The journal data is stored in `~/.journal-mcp/` with the following structure:
```
~/.journal-mcp/
├── tasks/          # Individual task files (JSON)
├── daily/          # Daily activity summaries  
├── weekly/         # Weekly summaries
└── one-on-ones/    # 1-on-1 meeting records
```

## MCP Tools

### Task Management
- `create_task` - Create new tasks with issue linking
- `add_task_entry` - Add timestamped entries to tasks
- `update_task_entry` - Modify existing entries
- `get_task` - Retrieve complete task history
- `list_tasks` - List tasks with filtering options
- `update_task_status` - Change task status (active/completed/paused/blocked)

### Time-based Views  
- `get_daily_log` - View all activity for a specific date
- `get_weekly_log` - View activity for a week

### 1-on-1 Management
- `create_one_on_one` - Record structured meeting notes
- `get_one_on_one_history` - Retrieve meeting history

### Search & Export
- `search_entries` - Search through all journal content
- `export_data` - Export to JSON, Markdown, or PDF

## Task Types

- **work** - Regular work tasks and bug fixes
- **learning** - Skill development and research projects  
- **personal** - Personal projects and side work
- **investigation** - Research and exploration tasks

## Example Workflow

1. **Start a new task:**
   ```json
   {
     "id": "MDU-1450",
     "title": "Fix payment timeout issues", 
     "type": "work",
     "tags": ["api", "debugging", "payment"],
     "issue_url": "https://github.com/company/repo/issues/1450",
     "priority": "high"
   }
   ```

2. **Add work entries throughout the day:**
   ```json
   {
     "timestamp": "2025-09-05T14:30:00Z",
     "content": "Started investigation - customer reports 8s timeouts in payment flow"
   }
   ```

3. **Update status when complete:**
   ```json
   {
     "status": "completed",
     "reason": "Fixed database query optimization, reduced to 200ms"
   }
   ```

## AI Integration

This MCP is designed to work with AI coding assistants to:

- **Automatically log work** - AI can create entries as you work
- **Suggest tasks** - Based on patterns and priorities  
- **Generate summaries** - For daily standups or reviews
- **Track learning** - Monitor skill development over time
- **Prepare 1-on-1s** - Surface relevant updates and blockers

## Development

### Prerequisites
- Go 1.21+
- Git

### Building
```bash
go build -o journal-mcp
```

### Testing  
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details
