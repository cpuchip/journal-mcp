package journal

import (
	"time"
	
	"github.com/google/go-github/v66/github"
)

// Service represents the core journal service
type Service struct {
	dataDir string
}

// Task represents a journal task with entries
type Task struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Type     string    `json:"type"` // work, learning, personal, investigation
	Tags     []string  `json:"tags"`
	Status   string    `json:"status"` // active, completed, paused, blocked
	Priority string    `json:"priority,omitempty"`
	IssueURL string    `json:"issue_url,omitempty"`
	IssueID  string    `json:"issue_id,omitempty"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	Entries  []Entry   `json:"entries"`
}

// Entry represents a timestamped journal entry
type Entry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
	Type      string    `json:"type,omitempty"` // log, status_change, completion, etc.
}

// OneOnOne represents structured meeting notes
type OneOnOne struct {
	Date     string    `json:"date"`
	Insights []string  `json:"insights,omitempty"`
	Todos    []string  `json:"todos,omitempty"`
	Feedback []string  `json:"feedback,omitempty"`
	Notes    string    `json:"notes,omitempty"`
	Created  time.Time `json:"created"`
}

// ImportResult represents the result of a data import operation
type ImportResult struct {
	TasksCreated     int      `json:"tasks_created"`
	EntriesAdded     int      `json:"entries_added"`
	DuplicatesSkipped int      `json:"duplicates_skipped"`
	Warnings         []string `json:"warnings,omitempty"`
	Summary          string   `json:"summary"`
}

// TaskRecommendation represents an AI-generated task recommendation
type TaskRecommendation struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Rationale   string  `json:"rationale"`
	Priority    string  `json:"priority"`
	Confidence  float64 `json:"confidence"`
	SuggestedTags []string `json:"suggested_tags,omitempty"`
}

// RecommendationsResult represents the result of task recommendations
type RecommendationsResult struct {
	Recommendations []TaskRecommendation `json:"recommendations"`
	AnalysisMetrics map[string]interface{} `json:"analysis_metrics"`
	Summary         string                 `json:"summary"`
}

// AnalyticsReport represents comprehensive analytics data
type AnalyticsReport struct {
	ReportType      string                 `json:"report_type"`
	TimePeriod      string                 `json:"time_period"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Summary         string                 `json:"summary"`
	TaskMetrics     TaskMetrics            `json:"task_metrics"`
	ProductivityMetrics ProductivityMetrics `json:"productivity_metrics"`
	PatternAnalysis PatternAnalysis        `json:"pattern_analysis"`
	Trends          []Trend                `json:"trends,omitempty"`
	Insights        []string               `json:"insights"`
}

// TaskMetrics represents task-related metrics
type TaskMetrics struct {
	TotalTasks       int                 `json:"total_tasks"`
	ByStatus         map[string]int      `json:"by_status"`
	ByType           map[string]int      `json:"by_type"`
	ByPriority       map[string]int      `json:"by_priority"`
	CompletionRate   float64             `json:"completion_rate"`
	AverageEntries   float64             `json:"average_entries_per_task"`
	TotalEntries     int                 `json:"total_entries"`
}

// ProductivityMetrics represents productivity-related metrics
type ProductivityMetrics struct {
	TasksCompletedPeriod  int     `json:"tasks_completed_period"`
	EntriesAddedPeriod    int     `json:"entries_added_period"`
	AverageTaskDuration   float64 `json:"average_task_duration_days"`
	MostProductiveType    string  `json:"most_productive_type"`
	ProductivityScore     float64 `json:"productivity_score"`
}

// PatternAnalysis represents pattern analysis data
type PatternAnalysis struct {
	MostFrequentType      string                `json:"most_frequent_type"`
	CommonTags            []string              `json:"common_tags"`
	WorkPatterns          map[string]int        `json:"work_patterns"`
	TimeToCompletion      map[string]float64    `json:"time_to_completion_by_type"`
}

// Trend represents a metric trend
type Trend struct {
	Metric    string  `json:"metric"`
	Direction string  `json:"direction"` // "up", "down", "stable"
	Change    float64 `json:"change"`
	Period    string  `json:"period"`
}

// DailyActivity represents daily activity data
type DailyActivity struct {
	Date  string             `json:"date"`
	Tasks map[string][]Entry `json:"tasks"` // task_id -> entries for that day
}

// GitHubService handles GitHub API interactions
type GitHubService struct {
	client *github.Client
	token  string
}

// GitHubSyncConfig holds configuration for GitHub sync
type GitHubSyncConfig struct {
	Token        string   `json:"token" yaml:"token"`
	Username     string   `json:"username" yaml:"username"`
	Repositories []string `json:"repositories" yaml:"repositories"`
	AutoSync     bool     `json:"auto_sync" yaml:"auto_sync"`
	SyncInterval int      `json:"sync_interval_minutes" yaml:"sync_interval_minutes"` // in minutes
}

// GitHubIssueMetadata stores additional GitHub-specific data
type GitHubIssueMetadata struct {
	IssueNumber int                        `json:"issue_number"`
	Repository  string                     `json:"repository"`
	State       string                     `json:"state"`
	Labels      []string                   `json:"labels"`
	Assignees   []string                   `json:"assignees"`
	Milestone   string                     `json:"milestone,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
	Comments    []GitHubIssueComment       `json:"comments,omitempty"`
	Events      []GitHubIssueEvent         `json:"events,omitempty"`
}

// GitHubIssueComment represents a comment on a GitHub issue
type GitHubIssueComment struct {
	ID        int64     `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GitHubIssueEvent represents an event on a GitHub issue
type GitHubIssueEvent struct {
	ID        int64     `json:"id"`
	Event     string    `json:"event"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
	Label     string    `json:"label,omitempty"`
}

// GitHubSyncResult represents the result of a GitHub sync operation
type GitHubSyncResult struct {
	TasksCreated    int      `json:"tasks_created"`
	TasksUpdated    int      `json:"tasks_updated"`
	IssuesProcessed int      `json:"issues_processed"`
	Errors          []string `json:"errors,omitempty"`
	Summary         string   `json:"summary"`
	LastSyncTime    time.Time `json:"last_sync_time"`
}

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