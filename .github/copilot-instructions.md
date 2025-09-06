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

### Phase 1: Core MCP Tool Completion
- [ ] **Implement Missing MCP Tools** 
  - `update_task_status` - Change task status (active/completed/paused/blocked)
  - `get_daily_log` - View all activity for a specific date  
  - `get_weekly_log` - View activity for a week
  - `create_one_on_one` - Record structured meeting notes
  - `get_one_on_one_history` - Retrieve meeting history
  - `search_entries` - Search through all journal content
  - `export_data` - Export to JSON, Markdown, or PDF

### Phase 2: Enhanced Features
- [ ] **Add AI-Assisted Todo Recommendations**
  - Analyze task patterns and suggest next actions
  - Priority-based task recommendations
  - Learning project progression suggestions

- [ ] **Import Capabilities**
  - Import existing diary/journal entries from various formats
  - Parse and structure existing data into task-based format
  - Support for common formats (txt, md, json, csv)

- [ ] **Advanced Search & Analytics**
  - Full-text search across all entries
  - Time-based analytics and insights
  - Task completion patterns and metrics

### Phase 3: Web Interface
- [ ] **Create Web Portal**
  - Modern React/Vue.js frontend
  - Real-time task viewing and editing
  - Visual timeline and calendar views
  - Search and filtering capabilities
  - Export functions through web UI

### Phase 4: Integration & Polish
- [ ] **Enhanced GitHub/Jira Integration**
  - Automatic task creation from GitHub issues
  - Sync task status with issue status
  - Pull in issue comments and updates

- [ ] **Testing & Quality**
  - Add comprehensive unit tests
  - Integration tests for MCP protocol
  - Performance optimization
  - Error handling improvements

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
# Test MCP server
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | .\journal-mcp.exe

# Create task example
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "create_task", "arguments": {"id": "TEST-123", "title": "Test task", "type": "work"}}}' | .\journal-mcp.exe
```

### Priority Order for Implementation
1. Complete missing MCP tools (highest impact)
2. Add comprehensive testing
3. Implement import capabilities  
4. Build web interface
5. Advanced integrations and analytics

**Repository**: https://github.com/cpuchip/journal-mcp  
**Status**: Core functionality complete, ready for feature expansion
