# Test Migration Guide for Phase 4 Reorganization

## 🎯 TESTING IMPACT ANALYSIS

### Current Test Structure
```
journal-mcp/
├── main_test.go          # 17 test functions (MCP server, tool registration)
├── journal_test.go       # 86 test functions (core journal functionality)  
├── phase3_test.go        # 30 test functions (GitHub, backup, config)
└── testdata/            # Test data files
    ├── sample_daily.json
    ├── sample_one_on_one.json
    └── sample_tasks.json
```

### After Reorganization
```
journal-mcp/
├── cmd/journal-mcp/
│   └── main_test.go      # Integration tests for main entry point
├── internal/
│   ├── server/
│   │   └── server_test.go      # MCP server setup tests (from main_test.go)
│   ├── journal/
│   │   └── service_test.go     # Core functionality tests (from journal_test.go)
│   ├── github/
│   │   └── client_test.go      # GitHub integration tests (from phase3_test.go)
│   ├── web/
│   │   └── server_test.go      # Web server tests
│   ├── config/
│   │   └── config_test.go      # Config tests (from phase3_test.go)
│   └── storage/
│       └── storage_test.go     # Storage and backup tests
└── testdata/               # Test data (moved up to root level)
```

## 🔧 IMPORT PATH CHANGES

### Before (Current)
```go
package main

import (
    "testing"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func TestToolRegistration(t *testing.T) {
    s := server.NewMCPServer("journal-mcp", "1.0.0", server.WithToolCapabilities(true))
    js := NewJournalService()
    registerTools(s, js)
    // test logic...
}
```

### After (Reorganized)
```go
package server

import (
    "testing"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/cpuchip/journal-mcp/internal/journal"
)

func TestToolRegistration(t *testing.T) {
    s := server.NewMCPServer("journal-mcp", "1.0.0", server.WithToolCapabilities(true))
    js := journal.NewService()
    RegisterTools(s, js)
    // test logic...
}
```

## 📋 MIGRATION CHECKLIST

### 1. Package Structure Changes
- [ ] Update package declarations in all test files
- [ ] Change function visibility (capitalize exported functions)
- [ ] Update import statements for internal packages
- [ ] Move test helper functions to appropriate packages

### 2. Test Data Access
- [ ] Update testdata file paths (may need `../../testdata/`)
- [ ] Ensure test data is accessible from new locations
- [ ] Consider copying testdata to each package if needed

### 3. Test Function Updates
- [ ] Change `NewJournalService()` to `journal.NewService()`
- [ ] Update `registerTools()` to `server.RegisterTools()`
- [ ] Fix any cross-package test dependencies
- [ ] Ensure test isolation is maintained

### 4. Build and CI Updates
- [ ] Update any build scripts that reference test files
- [ ] Verify `go test ./...` works from project root
- [ ] Check test coverage reporting still works
- [ ] Update any IDE test configurations

## ⚠️ POTENTIAL ISSUES

### Import Cycles
- **Risk**: Circular dependencies between packages
- **Solution**: Keep packages focused, use interfaces for decoupling
- **Watch for**: journal ↔ web, config ↔ storage dependencies

### Test Data Paths
- **Risk**: Relative paths to testdata/ may break
- **Solution**: Use build tags or embed testdata in test files
- **Alternative**: Copy necessary test files to each package

### Visibility Issues
- **Risk**: Private functions/types no longer accessible to tests
- **Solution**: Either make public or move tests to same package
- **Note**: Go allows tests in same package to access private members

## 🚀 RECOMMENDED APPROACH

### Step 1: Create New Structure (No Code Changes)
```bash
mkdir -p cmd/journal-mcp internal/{server,journal,github,web,config,storage}
```

### Step 2: Move Files Gradually
1. Start with config.go → internal/config/ (smallest impact)
2. Move github.go → internal/github/
3. Move webserver.go → internal/web/
4. Split journal.go → internal/journal/ + internal/storage/
5. Finally move main.go → cmd/journal-mcp/

### Step 3: Update Tests After Each Move
1. Move related test file to same package
2. Update package declaration and imports
3. Run tests to ensure they still pass
4. Fix any issues before moving to next file

### Step 4: Verify Complete Test Suite
```bash
# This should continue to work after reorganization
go test -v ./...

# This should show same coverage percentage
go test -cover ./...

# Integration test - MCP server should still work
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}' | go run ./cmd/journal-mcp
```

## 💡 TESTING BEST PRACTICES FOR NEW STRUCTURE

### Package-Level Tests
Each package should have comprehensive tests for its public API:

```go
// internal/journal/service_test.go
package journal

func TestService_CreateTask(t *testing.T) {
    s := NewService()
    // Test internal service methods directly
}

// internal/server/server_test.go  
package server

func TestMCPIntegration(t *testing.T) {
    // Test full MCP protocol integration
}
```

### Integration Tests
Keep integration tests at the cmd level:

```go
// cmd/journal-mcp/main_test.go
package main

func TestFullWorkflow(t *testing.T) {
    // Test complete workflows end-to-end
}
```

This reorganization will result in better test organization, clearer separation of concerns, and easier maintenance while preserving all existing test coverage.
