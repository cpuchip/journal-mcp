<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->
- [x] Verify that the copilot-instructions.md file in the .github directory is created.

- [x] Clarify Project Requirements
	<!-- Go-based MCP server using mcp-go library for journaling system with task-based logging, web UI, and structured data storage -->

- [x] Scaffold the Project
	<!-- Go MCP project structure created with main.go, journal.go, and supporting files -->

- [x] Customize the Project
	<!-- Project customized for journal MCP with task-based logging, structured data, and MCP tool implementations -->

- [x] Install Required Extensions
	<!-- No specific extensions required for Go MCP project -->

- [x] Compile the Project
	<!-- Successfully compiled journal-mcp.exe binary, all dependencies resolved -->

- [x] Create and Run Task
	<!-- Not needed - this is an MCP server, not a traditional VS Code project -->

- [x] Launch the Project
	<!-- Successfully tested MCP server with JSON-RPC protocol, fully functional -->

- [x] Ensure Documentation is Complete
	<!-- README.md complete, GitHub repository created and pushed -->

## 🎯 REMAINING TASKS FOR CODING AGENT

### ✅ Phase 1: Core MCP Tool Completion (COMPLETED)
- [x] **All Missing MCP Tools Implemented** 
  - [x] `update_task_status` - Change task status (active/completed/paused/blocked)
  - [x] `get_daily_log` - View all activity for a specific date  
  - [x] `get_weekly_log` - View activity for a week
  - [x] `create_one_on_one` - Record structured meeting notes
  - [x] `get_one_on_one_history` - Retrieve meeting history
  - [x] `search_entries` - Search through all journal content
  - [x] `export_data` - Export to JSON, Markdown, or CSV

- [x] **Polish & Bug Fixes Completed**
  - [x] Complete date filtering implementation (`date_from`, `date_to` parameters)
  - [x] Enhanced error handling with user-friendly messages
  - [x] Date format validation across all date-accepting tools
  - [x] Performance optimization with pagination support (`limit`, `offset`)
  - [x] Comprehensive unit tests (70.2% coverage, 46 test cases)
  - [x] All edge cases properly handled

### ✅ Phase 2: Enhanced Features (COMPLETED)
- [x] **Add AI-Assisted Todo Recommendations**
  - [x] Analyze task patterns and suggest next actions
  - [x] Priority-based task recommendations
  - [x] Learning project progression suggestions

- [x] **Import Capabilities**
  - [x] Import existing diary/journal entries from various formats
  - [x] Parse and structure existing data into task-based format
  - [x] Support for common formats (txt, md, json, csv)

- [x] **Advanced Search & Analytics**
  - [x] Full-text search across all entries
  - [x] Time-based analytics and insights
  - [x] Task completion patterns and metrics

### ✅ Phase 3: GitHub Integration and Web Foundation (COMPLETED)
- [x] **Enhanced GitHub/Jira Integration**
  - [x] Automatic task creation from GitHub issues
  - [x] Sync task status with issue status
  - [x] Pull in issue comments and updates
  - [x] OAuth-based GitHub authentication

- [x] **Web Interface Foundation**
  - [x] REST API with 25+ endpoints
  - [x] WebSocket support for real-time updates
  - [x] CORS middleware and proper HTTP handling
  - [x] Demo HTML interface

- [x] **Data Management**
  - [x] Backup/restore with compression
  - [x] Configuration management (YAML)
  - [x] Migration framework for future enhancements

### 🚀 Phase 4: Production-Ready Application (ACTIVE PRIORITY)
- [ ] **Professional Go Project Structure** (HIGH PRIORITY)
  - [ ] Reorganize to standard Go layout: cmd/, internal/, pkg/, frontend/
  - [ ] Move main.go to cmd/journal-mcp/main.go
  - [ ] Split large files into focused packages (journal/, web/, config/, github/)
  - [ ] Update all import paths and ensure 133+ tests still pass
  - [ ] Add Makefile for build automation

- [ ] **Embedded Web Interface** (HIGH PRIORITY)
  - [ ] Build modern Vue.js SPA with offline-first design
  - [ ] Bundle ALL dependencies locally (no external CDNs)
  - [ ] Embed built frontend in Go binary using embed.FS
  - [ ] Comprehensive UI for all 22 MCP tools
  - [ ] Real-time updates via WebSocket integration

- [ ] **Enhanced User Experience** (MEDIUM PRIORITY)
  - [ ] Dashboard: task summary, recent activity, analytics charts
  - [ ] Advanced task management: drag-drop, bulk operations, filters
  - [ ] Timeline and calendar views with visual density
  - [ ] Modern responsive design (desktop + mobile)
  - [ ] Keyboard shortcuts and accessibility (WCAG 2.1 AA)

- [ ] **Production Polish** (MEDIUM PRIORITY)
  - [ ] Performance optimization and loading states
  - [ ] Comprehensive error handling and user feedback
  - [ ] Documentation updates and help system
  - [ ] Single binary deployment with all assets embedded

## 🛠️ TECHNICAL NOTES FOR CODING AGENT

### Current Architecture
- **MCP Server**: Uses mcp-go v0.39.1 library
- **Data Storage**: JSON files in `~/.journal-mcp/` directory
- **Transport**: Stdio protocol for MCP communication
- **Structure**: Task-centric with temporal views

### Key Implementation Details
- Function signatures use `(ctx context.Context, request mcp.CallToolRequest)` 
- Use `request.RequireString()` and `request.GetString()` helper methods
- Return `mcp.NewToolResultText()` or `mcp.NewToolResultError()`
- Tool registration uses `mcp.NewTool()` with option functions

### Development Environment
- Go 1.23.2
- Windows PowerShell development environment
- GitHub repository: https://github.com/cpuchip/journal-mcp
- Local data storage in user home directory

### Testing Commands
```powershell
# Run all tests with coverage
go test -cover ./...

# Test MCP server
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | .\journal-mcp.exe

# Create task example
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_task", "arguments": {"id": "TEST-123", "title": "Test task", "type": "work"}}}' | .\journal-mcp.exe

# Test pagination
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "list_tasks", "arguments": {"limit": "5", "offset": "0"}}}' | .\journal-mcp.exe

# Test date filtering
echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "list_tasks", "arguments": {"date_from": "2025-01-01", "date_to": "2025-12-31"}}}' | .\journal-mcp.exe
```

### Priority Order for Implementation
1. ✅ Complete missing MCP tools (COMPLETED - all 11 tools implemented)
2. ✅ Add comprehensive testing (COMPLETED - 70.2% coverage achieved)
3. ✅ Enhanced filtering and pagination (COMPLETED)
4. ✅ Import capabilities (COMPLETED)
5. ✅ AI-assisted recommendations (COMPLETED)
6. ✅ GitHub integration and web interface foundation (COMPLETED)
7. **PROJECT REORGANIZATION AND ENHANCED WEB INTERFACE (CURRENT PRIORITY)**

### 🚀 Phase 4: Production-Ready Application (ACTIVE)
- [ ] **Professional Go Project Structure**
  - Reorganize files: cmd/journal-mcp/main.go, internal/* packages
  - Update all import paths and ensure tests pass
  - Follow Go community standards and best practices

- [ ] **Embedded Web Interface**
  - Build modern Vue.js SPA with offline-first design
  - Bundle all dependencies (no external CDNs)
  - Embed frontend in Go binary using embed.FS
  - Comprehensive UI for all MCP tools

- [ ] **Enhanced User Experience**
  - Dashboard with task summary, recent activity, analytics
  - Real-time updates via WebSocket
  - Modern, responsive design (desktop + mobile)
  - Advanced features: drag-drop, bulk operations, keyboard shortcuts

**Repository**: https://github.com/cpuchip/journal-mcp  
**Status**: Phases 1-3 complete! All core functionality implemented. Phase 4 focus: Professional structure + embedded web interface.  
**Documentation**: See PHASE4_INSTRUCTIONS.md for complete reorganization and enhancement plan.
