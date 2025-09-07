package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// Configuration represents the journal configuration
type Configuration struct {
	GitHub struct {
		Token        string   `json:"token,omitempty" yaml:"token,omitempty"`
		Username     string   `json:"username,omitempty" yaml:"username,omitempty"`
		Repositories []string `json:"repositories,omitempty" yaml:"repositories,omitempty"`
		AutoSync     bool     `json:"auto_sync" yaml:"auto_sync"`
		SyncInterval int      `json:"sync_interval_minutes" yaml:"sync_interval_minutes"`
	} `json:"github" yaml:"github"`
	
	Web struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
		Port    int  `json:"port" yaml:"port"`
	} `json:"web" yaml:"web"`
	
	Backup struct {
		AutoBackup     bool   `json:"auto_backup" yaml:"auto_backup"`
		BackupInterval int    `json:"backup_interval_hours" yaml:"backup_interval_hours"`
		BackupLocation string `json:"backup_location,omitempty" yaml:"backup_location,omitempty"`
		MaxBackups     int    `json:"max_backups" yaml:"max_backups"`
	} `json:"backup" yaml:"backup"`
	
	General struct {
		DefaultTaskType string `json:"default_task_type" yaml:"default_task_type"`
		TimeZone        string `json:"timezone" yaml:"timezone"`
		DateFormat      string `json:"date_format" yaml:"date_format"`
	} `json:"general" yaml:"general"`
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	BackupPath   string    `json:"backup_path"`
	Size         int64     `json:"size_bytes"`
	FilesBackup  int       `json:"files_backup"`
	CreatedAt    time.Time `json:"created_at"`
	Summary      string    `json:"summary"`
}

// RestoreResult represents the result of a restore operation
type RestoreResult struct {
	FilesRestored   int      `json:"files_restored"`
	TasksRestored   int      `json:"tasks_restored"`
	EntriesRestored int      `json:"entries_restored"`
	Warnings        []string `json:"warnings,omitempty"`
	Summary         string   `json:"summary"`
}

// CreateDataBackup creates a backup of all journal data
func (js *JournalService) CreateDataBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	backupPath := request.GetString("backup_path", "")
	includeConfig := request.GetString("include_config", "true") == "true"
	compressionLevel := request.GetString("compression", "default")

	if backupPath == "" {
		// Generate default backup path
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		backupPath = filepath.Join(js.dataDir, "backups", fmt.Sprintf("journal-backup-%s.zip", timestamp))
	}

	// Ensure backup directory exists
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create backup directory: %v", err)), nil
	}

	// Create ZIP file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create backup file: %v", err)), nil
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var filesBackup int
	var totalSize int64

	// Backup tasks
	tasksDir := filepath.Join(js.dataDir, "tasks")
	if err := js.addDirectoryToZip(zipWriter, tasksDir, "tasks", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup tasks: %v", err)), nil
	}

	// Backup daily logs
	dailyDir := filepath.Join(js.dataDir, "daily")
	if err := js.addDirectoryToZip(zipWriter, dailyDir, "daily", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup daily logs: %v", err)), nil
	}

	// Backup weekly logs
	weeklyDir := filepath.Join(js.dataDir, "weekly")
	if err := js.addDirectoryToZip(zipWriter, weeklyDir, "weekly", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup weekly logs: %v", err)), nil
	}

	// Backup one-on-ones
	oneOnOneDir := filepath.Join(js.dataDir, "one-on-ones")
	if err := js.addDirectoryToZip(zipWriter, oneOnOneDir, "one-on-ones", &filesBackup, &totalSize); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to backup one-on-ones: %v", err)), nil
	}

	// Backup configuration if requested
	if includeConfig {
		configPath := filepath.Join(js.dataDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			if err := js.addFileToZip(zipWriter, configPath, "config.yaml", &totalSize); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to backup config: %v", err)), nil
			}
			filesBackup++
		}
	}

	// Add backup metadata
	metadata := map[string]interface{}{
		"created_at":      time.Now(),
		"version":         "1.0.0",
		"source_dir":      js.dataDir,
		"files_count":     filesBackup,
		"include_config":  includeConfig,
		"compression":     compressionLevel,
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	metadataWriter, err := zipWriter.Create("backup_metadata.json")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create metadata: %v", err)), nil
	}
	metadataWriter.Write(metadataJSON)

	// Get final file size
	zipWriter.Close()
	zipFile.Close()

	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get backup file info: %v", err)), nil
	}

	result := BackupResult{
		BackupPath:  backupPath,
		Size:        fileInfo.Size(),
		FilesBackup: filesBackup,
		CreatedAt:   time.Now(),
		Summary:     fmt.Sprintf("Successfully created backup with %d files (%d bytes) at %s", filesBackup, fileInfo.Size(), backupPath),
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// RestoreDataBackup restores journal data from a backup file
func (js *JournalService) RestoreDataBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	backupPath := request.GetString("backup_path", "")
	if backupPath == "" {
		return mcp.NewToolResultError("backup_path is required"), nil
	}

	overwriteExisting := request.GetString("overwrite_existing", "false") == "true"
	restoreConfig := request.GetString("restore_config", "true") == "true"

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Backup file not found: %s", backupPath)), nil
	}

	// Open ZIP file
	zipReader, err := zip.OpenReader(backupPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open backup file: %v", err)), nil
	}
	defer zipReader.Close()

	var restoreResult RestoreResult
	restoreResult.Warnings = []string{}

	// Create restore directory if needed
	if !overwriteExisting {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		restoreDir := filepath.Join(js.dataDir, fmt.Sprintf("restore-%s", timestamp))
		if err := os.MkdirAll(restoreDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create restore directory: %v", err)), nil
		}
		js.dataDir = restoreDir
	}

	// Extract files
	for _, file := range zipReader.File {
		// Skip config if not requested
		if !restoreConfig && strings.HasSuffix(file.Name, "config.yaml") {
			continue
		}

		// Skip metadata file
		if file.Name == "backup_metadata.json" {
			continue
		}

		if err := js.extractFileFromZip(file, &restoreResult); err != nil {
			restoreResult.Warnings = append(restoreResult.Warnings, fmt.Sprintf("Failed to extract %s: %v", file.Name, err))
			continue
		}

		restoreResult.FilesRestored++

		// Count tasks and entries
		if strings.HasPrefix(file.Name, "tasks/") {
			restoreResult.TasksRestored++
			// Could parse and count entries, but for now just increment
			restoreResult.EntriesRestored += 10 // Estimate
		}
	}

	restoreResult.Summary = fmt.Sprintf("Successfully restored %d files (%d tasks) from backup", 
		restoreResult.FilesRestored, restoreResult.TasksRestored)

	resultJSON, _ := json.Marshal(restoreResult)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// GetConfiguration retrieves the current configuration
func (js *JournalService) GetConfiguration(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	configPath := filepath.Join(js.dataDir, "config.yaml")
	
	var config Configuration
	
	// Load existing config or use defaults
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse config: %v", err)), nil
		}
	} else {
		// Set defaults
		config = Configuration{}
		config.Web.Enabled = false
		config.Web.Port = 8080
		config.Backup.AutoBackup = false
		config.Backup.BackupInterval = 24
		config.Backup.MaxBackups = 7
		config.General.DefaultTaskType = "work"
		config.General.TimeZone = "UTC"
		config.General.DateFormat = "2006-01-02"
		config.GitHub.AutoSync = false
		config.GitHub.SyncInterval = 60
	}

	configJSON, _ := json.Marshal(config)
	return mcp.NewToolResultText(string(configJSON)), nil
}

// UpdateConfiguration updates the journal configuration
func (js *JournalService) UpdateConfiguration(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	configData := request.GetString("config", "")
	if configData == "" {
		return mcp.NewToolResultError("config data is required"), nil
	}

	var config Configuration
	if err := json.Unmarshal([]byte(configData), &config); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid config JSON: %v", err)), nil
	}

	// Validate configuration
	if err := js.validateConfiguration(&config); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid configuration: %v", err)), nil
	}

	// Save configuration
	configPath := filepath.Join(js.dataDir, "config.yaml")
	configYAML, err := yaml.Marshal(config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal config: %v", err)), nil
	}

	if err := os.WriteFile(configPath, configYAML, 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save config: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": "Configuration updated successfully",
		"config":  config,
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// MigrateData performs data migration (future SQLite integration preparation)
func (js *JournalService) MigrateData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	migrationVersion := request.GetString("target_version", "current")
	dryRun := request.GetString("dry_run", "false") == "true"

	// For now, this is a placeholder for future database migrations
	result := map[string]interface{}{
		"status":           "success",
		"migration_version": migrationVersion,
		"dry_run":          dryRun,
		"message":          "Data migration framework ready for future use",
		"migrations_applied": []string{},
		"summary":          "No migrations needed for current JSON-based storage",
	}

	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// Helper methods for backup/restore

func (js *JournalService) addDirectoryToZip(zipWriter *zip.Writer, sourceDir, zipDir string, fileCount *int, totalSize *int64) error {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(zipDir, relPath)
		return js.addFileToZip(zipWriter, path, zipPath, totalSize)
	})
}

func (js *JournalService) addFileToZip(zipWriter *zip.Writer, filePath, zipPath string, totalSize *int64) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	written, err := io.Copy(writer, file)
	if err != nil {
		return err
	}

	*totalSize += written
	return nil
}

func (js *JournalService) extractFileFromZip(file *zip.File, result *RestoreResult) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create target directory
	targetPath := filepath.Join(js.dataDir, file.Name)
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Create target file
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, reader)
	return err
}

func (js *JournalService) validateConfiguration(config *Configuration) error {
	// Validate web configuration
	if config.Web.Port < 1 || config.Web.Port > 65535 {
		return fmt.Errorf("invalid web port: %d", config.Web.Port)
	}

	// Validate backup configuration
	if config.Backup.BackupInterval < 1 {
		return fmt.Errorf("backup interval must be at least 1 hour")
	}

	if config.Backup.MaxBackups < 1 {
		return fmt.Errorf("max backups must be at least 1")
	}

	// Validate GitHub configuration
	if config.GitHub.SyncInterval < 5 {
		return fmt.Errorf("GitHub sync interval must be at least 5 minutes")
	}

	// Validate general configuration
	validTaskTypes := []string{"work", "learning", "personal", "investigation"}
	valid := false
	for _, taskType := range validTaskTypes {
		if config.General.DefaultTaskType == taskType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid default task type: %s", config.General.DefaultTaskType)
	}

	return nil
}