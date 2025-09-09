package servers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGitHubSyncWithGitHub(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing github_token",
			args: map[string]interface{}{
				"username": "testuser",
			},
			expectError: true,
			errorMsg:    "github_token is required",
		},
		{
			name: "missing username",
			args: map[string]interface{}{
				"github_token": "test-token",
			},
			expectError: true,
			errorMsg:    "username is required",
		},
		{
			name: "valid sync request",
			args: map[string]interface{}{
				"github_token": "test-token",
				"username":     "testuser",
				"repositories": []string{"owner/repo"},
				"create_tasks": "true",
			},
			expectError: true, // Will fail due to invalid token, but validates structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CreateMockRequest(tt.args)
			result, err := js.SyncWithGitHub(ctx, request)

			if tt.expectError {
				if err == nil && !result.IsError {
					t.Errorf("Expected error but got success")
				}
				if result.IsError {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						if textContent.Text != tt.errorMsg {
							t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, textContent.Text)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
			}
		})
	}
}

func TestCreateDataBackup(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	// Create some test data first
	createTestTask(t, js, "backup-test-1", "Test Task 1", "work")
	createTestTask(t, js, "backup-test-2", "Test Task 2", "learning")

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
	}{
		{
			name: "create backup with default settings",
			args: map[string]interface{}{
				"include_config": "true",
				"compression":    "default",
			},
			expectError: false,
		},
		{
			name: "create backup without config",
			args: map[string]interface{}{
				"include_config": "false",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CreateMockRequest(tt.args)
			result, err := js.CreateDataBackup(ctx, request)

			if tt.expectError {
				if err == nil && !result.IsError {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsError {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						t.Errorf("Expected success but got error: %s", textContent.Text)
					}
				}

				// Verify backup result structure
				if !result.IsError && len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						var backupResult BackupResult
						if err := json.Unmarshal([]byte(textContent.Text), &backupResult); err != nil {
							t.Errorf("Failed to parse backup result: %v", err)
						} else {
							if backupResult.FilesBackup == 0 {
								t.Errorf("Expected some files to be backed up")
							}
							if backupResult.Size == 0 {
								t.Errorf("Expected backup file to have non-zero size")
							}
						}
					}
				}
			}
		})
	}
}

func TestGetConfiguration(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	request := CreateMockRequest(map[string]interface{}{})
	result, err := js.GetConfiguration(ctx, request)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("Expected success but got error")
	}

	// Verify configuration structure
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			var config Configuration
			if err := json.Unmarshal([]byte(textContent.Text), &config); err != nil {
				t.Errorf("Failed to parse configuration: %v", err)
			} else {
				// Verify default values
				if config.Web.Port == 0 {
					t.Errorf("Expected default web port to be set")
				}
				if config.General.DefaultTaskType == "" {
					t.Errorf("Expected default task type to be set")
				}
			}
		}
	}
}

func TestUpdateConfiguration(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		configJSON  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing config",
			configJSON:  "",
			expectError: true,
			errorMsg:    "config data is required",
		},
		{
			name:        "invalid JSON",
			configJSON:  "{invalid json",
			expectError: true,
			errorMsg:    "Invalid config JSON",
		},
		{
			name: "valid configuration",
			configJSON: `{
				"web": {
					"enabled": true,
					"port": 9090
				},
				"general": {
					"default_task_type": "work",
					"timezone": "America/New_York"
				}
			}`,
			expectError: false,
		},
		{
			name: "invalid port",
			configJSON: `{
				"web": {
					"port": 99999
				},
				"general": {
					"default_task_type": "work"
				}
			}`,
			expectError: true,
			errorMsg:    "Invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{}
			if tt.configJSON != "" {
				args["config"] = tt.configJSON
			}

			request := CreateMockRequest(args)
			result, err := js.UpdateConfiguration(ctx, request)

			if tt.expectError {
				if err == nil && !result.IsError {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsError {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						t.Errorf("Expected success but got error: %s", textContent.Text)
					}
				}
			}
		})
	}
}

func TestPullIssueUpdates(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing github_token",
			args: map[string]interface{}{
				"task_id": "test-task",
			},
			expectError: true,
			errorMsg:    "github_token is required",
		},
		{
			name: "invalid since timestamp",
			args: map[string]interface{}{
				"github_token": "test-token",
				"since":        "invalid-timestamp",
			},
			expectError: true,
			errorMsg:    "Invalid since timestamp format",
		},
		{
			name: "valid request with timestamp",
			args: map[string]interface{}{
				"github_token": "test-token",
				"since":        "2023-01-01T00:00:00Z",
			},
			expectError: false, // Will fail due to no tasks with GitHub issues, but structure is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CreateMockRequest(tt.args)
			result, err := js.PullIssueUpdates(ctx, request)

			if tt.expectError {
				if err == nil && !result.IsError {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// For this test, we expect success even with no tasks to update
				if result.IsError {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						t.Errorf("Expected success but got error: %s", textContent.Text)
					}
				}
			}
		})
	}
}

func TestCreateTaskFromGitHubIssue(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing github_token",
			args: map[string]interface{}{
				"issue_url": "https://github.com/owner/repo/issues/123",
			},
			expectError: true,
			errorMsg:    "github_token is required",
		},
		{
			name: "missing issue_url",
			args: map[string]interface{}{
				"github_token": "test-token",
			},
			expectError: true,
			errorMsg:    "issue_url is required",
		},
		{
			name: "invalid issue_url",
			args: map[string]interface{}{
				"github_token": "test-token",
				"issue_url":    "not-a-github-url",
			},
			expectError: true,
			errorMsg:    "Invalid GitHub URL",
		},
		{
			name: "valid request structure",
			args: map[string]interface{}{
				"github_token": "test-token",
				"issue_url":    "https://github.com/owner/repo/issues/123",
				"type":         "work",
				"priority":     "high",
			},
			expectError: true, // Will fail due to invalid token, but validates structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CreateMockRequest(tt.args)
			result, err := js.CreateTaskFromGitHubIssue(ctx, request)

			if tt.expectError {
				if err == nil && !result.IsError {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.IsError {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						t.Errorf("Expected success but got error: %s", textContent.Text)
					}
				}
			}
		})
	}
}

func TestMigrateData(t *testing.T) {
	js, _ := CreateTestJournalService(t)
	ctx := context.Background()

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "default migration",
			args: map[string]interface{}{},
		},
		{
			name: "dry run migration",
			args: map[string]interface{}{
				"dry_run": "true",
			},
		},
		{
			name: "specific version migration",
			args: map[string]interface{}{
				"target_version": "2.0.0",
				"dry_run":        "false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CreateMockRequest(tt.args)
			result, err := js.MigrateData(ctx, request)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result.IsError {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					t.Errorf("Expected success but got error: %s", textContent.Text)
				}
			}

			// Verify migration result structure
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
					var migrationResult map[string]interface{}
					if err := json.Unmarshal([]byte(textContent.Text), &migrationResult); err != nil {
						t.Errorf("Failed to parse migration result: %v", err)
					} else {
						if migrationResult["status"] != "success" {
							t.Errorf("Expected successful migration")
						}
					}
				}
			}
		})
	}
}

// Helper function to create test tasks
func createTestTask(t *testing.T, js *JournalService, id, title, taskType string) {
	args := map[string]interface{}{
		"id":    id,
		"title": title,
		"type":  taskType,
	}
	request := CreateMockRequest(args)
	_, err := js.CreateTask(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
}
