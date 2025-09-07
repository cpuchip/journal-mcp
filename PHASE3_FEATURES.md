# Journal MCP - Phase 3 Features

This document describes the Phase 3 enhancements to the Journal MCP system, including enhanced GitHub integration, web interface foundation, and data management improvements.

## Phase 3 Features Overview

### 1. Enhanced GitHub Integration

The system now provides comprehensive GitHub integration capabilities:

#### GitHub Sync Tools

- **`sync_with_github`** - Automatically sync assigned GitHub issues with tasks
  ```json
  {
    "github_token": "your-github-token",
    "username": "your-username",
    "repositories": ["owner/repo1", "owner/repo2"],
    "create_tasks": "true",
    "update_existing": "true"
  }
  ```

- **`pull_issue_updates`** - Pull latest comments and events for tracked GitHub issues
  ```json
  {
    "github_token": "your-github-token",
    "task_id": "optional-specific-task",
    "since": "2024-01-01T00:00:00Z"
  }
  ```

- **`create_task_from_github_issue`** - Create a new task from a GitHub issue URL
  ```json
  {
    "github_token": "your-github-token",
    "issue_url": "https://github.com/owner/repo/issues/123",
    "type": "work",
    "priority": "high"
  }
  ```

#### GitHub Features

- **Automatic Task Creation**: Creates tasks automatically from assigned GitHub issues
- **Bidirectional Sync**: Updates task status when GitHub issues are closed
- **Comment Integration**: Pulls issue comments as task entries
- **Event Tracking**: Tracks GitHub issue events (labels, assignments, etc.)
- **Metadata Storage**: Stores GitHub-specific metadata (labels, assignees, milestones)

### 2. Web Interface Foundation

#### REST API Server

The system now includes a comprehensive REST API that can run alongside or instead of the MCP server:

```bash
# Run web server only
./journal-mcp --web

# Run both MCP and web server
./journal-mcp --dual

# Default MCP-only mode
./journal-mcp
```

#### API Endpoints

**Task Management**
- `GET /api/tasks` - List all tasks with filtering
- `POST /api/tasks` - Create a new task
- `GET /api/tasks/{id}` - Get specific task
- `PUT /api/tasks/{id}` - Update task (planned)
- `DELETE /api/tasks/{id}` - Delete task (planned)
- `POST /api/tasks/{id}/entries` - Add task entry
- `PUT /api/tasks/{id}/status` - Update task status

**Search & Analytics**
- `GET /api/search?q=query&task_type=work` - Search entries
- `GET /api/analytics/overview` - Get analytics overview
- `GET /api/analytics/report?type=productivity&period=month` - Get detailed reports

**Data Export**
- `GET /api/export?format=json&date_from=2024-01-01` - Export data

**Logs**
- `GET /api/logs/daily/{date}` - Get daily activity log
- `GET /api/logs/weekly/{date}` - Get weekly activity log

**One-on-Ones**
- `GET /api/one-on-ones` - List one-on-one meetings
- `POST /api/one-on-ones` - Create meeting record

**GitHub Integration**
- `POST /api/github/sync` - Sync with GitHub issues
- `POST /api/github/pull-updates` - Pull latest issue updates
- `POST /api/github/create-task` - Create task from GitHub issue

**Data Management**
- `POST /api/backup` - Create data backup
- `POST /api/restore` - Restore from backup
- `GET /api/config` - Get configuration
- `PUT /api/config` - Update configuration

**Real-time Updates**
- `WS /api/ws` - WebSocket endpoint for real-time updates

**Documentation**
- `GET /api/docs` - OpenAPI documentation
- `GET /api/health` - Health check

#### Frontend Architecture Planning

The API is designed to support modern frontend frameworks:

**Vue.js Integration**
- RESTful API design with consistent JSON responses
- WebSocket support for real-time updates
- CORS enabled for cross-origin requests
- Comprehensive error handling

**Planned UI Components**
- Task dashboard with filtering and search
- Real-time activity timeline
- Analytics and reporting dashboard
- GitHub issue integration panel
- Configuration management interface

### 3. Data Management Improvements

#### Backup & Restore System

- **`create_data_backup`** - Create comprehensive data backups
  ```json
  {
    "backup_path": "/path/to/backup.zip",
    "include_config": "true",
    "compression": "default"
  }
  ```

- **`restore_data_backup`** - Restore from backup files
  ```json
  {
    "backup_path": "/path/to/backup.zip",
    "overwrite_existing": "false",
    "restore_config": "true"
  }
  ```

#### Configuration Management

- **`get_configuration`** - Retrieve current configuration
- **`update_configuration`** - Update system configuration

**Configuration Structure**
```yaml
github:
  token: "your-github-token"
  username: "your-username"
  repositories: ["owner/repo1", "owner/repo2"]
  auto_sync: true
  sync_interval_minutes: 60

web:
  enabled: true
  port: 8080

backup:
  auto_backup: true
  backup_interval_hours: 24
  backup_location: "/path/to/backups"
  max_backups: 7

general:
  default_task_type: "work"
  timezone: "UTC"
  date_format: "2006-01-02"
```

#### Data Migration Framework

- **`migrate_data`** - Future-proofing for SQLite integration
  ```json
  {
    "target_version": "2.0.0",
    "dry_run": "false"
  }
  ```

## Installation & Setup

### Prerequisites
- Go 1.23.2+
- GitHub Personal Access Token (for GitHub integration)

### Build & Run

```bash
# Install dependencies
go mod tidy

# Build the application
go build -o journal-mcp

# Run MCP server only (default)
./journal-mcp

# Run web server only
./journal-mcp --web

# Run both MCP and web server
./journal-mcp --dual
```

### GitHub Integration Setup

1. Create a GitHub Personal Access Token:
   - Go to GitHub Settings → Developer settings → Personal access tokens
   - Generate a new token with `repo` and `read:user` scopes

2. Configure the system:
   ```bash
   # Update configuration with your GitHub token
   echo '{"github": {"token": "your-token", "username": "your-username"}}' | \
   ./journal-mcp tools call update_configuration --config
   ```

3. Sync with GitHub:
   ```bash
   # Sync assigned issues
   ./journal-mcp tools call sync_with_github --github_token "your-token" --username "your-username"
   ```

## API Usage Examples

### Create a Task via REST API

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "id": "PROJ-123",
    "title": "Implement new feature",
    "type": "work",
    "priority": "high",
    "tags": ["backend", "api"]
  }'
```

### Search Entries

```bash
curl "http://localhost:8080/api/search?q=bug%20fix&task_type=work&date_from=2024-01-01"
```

### Get Analytics

```bash
curl "http://localhost:8080/api/analytics/report?type=productivity&period=month"
```

### Sync with GitHub

```bash
curl -X POST http://localhost:8080/api/github/sync \
  -H "Content-Type: application/json" \
  -d '{
    "github_token": "your-token",
    "username": "your-username",
    "create_tasks": "true"
  }'
```

### Create Data Backup

```bash
curl -X POST http://localhost:8080/api/backup \
  -H "Content-Type: application/json" \
  -d '{
    "include_config": "true",
    "compression": "default"
  }'
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test suites
go test -run TestGitHub ./...
go test -run TestBackup ./...
```

### API Development

The web server provides extensive APIs for frontend development:

1. **Development Server**: Use `--web` mode for frontend development
2. **API Documentation**: Access `/api/docs` for OpenAPI specification
3. **Health Checks**: Use `/api/health` for monitoring
4. **WebSocket**: Connect to `/api/ws` for real-time updates

### Database Migration (Future)

The migration framework is ready for future SQLite integration:

```bash
# Prepare for migration (dry run)
./journal-mcp tools call migrate_data --target_version "2.0.0" --dry_run "true"

# Perform actual migration
./journal-mcp tools call migrate_data --target_version "2.0.0"
```

## Security Considerations

1. **GitHub Tokens**: Store securely, use environment variables in production
2. **Web Server**: Configure proper CORS policies for production
3. **Backup Files**: Encrypt sensitive backup data
4. **API Access**: Consider implementing authentication for production use

## Roadmap

### Phase 4 (Planned)
- Complete web frontend implementation
- Enhanced analytics and reporting
- Real-time collaboration features
- Advanced GitHub webhook integration
- Database migration to SQLite
- Cloud backup integration (OneDrive, etc.)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details