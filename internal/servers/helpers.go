package servers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateTestJournalService creates a temporary journal service for testing
// This function is exported so it can be used by other packages' tests
func CreateTestJournalService(t *testing.T) (*JournalService, string) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create subdirectories
	os.MkdirAll(filepath.Join(tempDir, "tasks"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "daily"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "weekly"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "one-on-ones"), 0755)

	return &JournalService{DataDir: tempDir}, tempDir
}

// CreateMockRequest creates a mock MCP request for testing
// This function is exported so it can be used by other packages' tests
func CreateMockRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}
