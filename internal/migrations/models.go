package migrations

import "database/sql"

// Config holds the application configuration
type Config struct {
	Paths struct {
		ConfigFile   string `json:"config_file"`
		ModelDir     string `json:"model_dir"`
		StateFile    string `json:"state_file"`
		MigrationDir string `json:"migration_dir"`
		Archived_dir string `json:"archived_dir"`
		HistoryDir   string `json:"history_dir"`
	} `json:"paths"`

	Database struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
		SSLMode  string `json:"sslmode"`
	} `json:"database"`

	Settings struct {
		VersionPrefixLength int `json:"version_prefix_length"`
	} `json:"settings"`

	Version int `json:"version"`
	Models  []struct {
		GoFile     string `json:"go_file"`
		StructName string `json:"struct_name"`
		TableName  string `json:"table_name"`
		OutFile    string `json:"out_file"`
	} `json:"models"`

	// ... existing fields ...
	StrictMode bool `json:"strict_mode"` // New flag
}

// ModelConfig represents a single model configuration
type ModelConfig struct {
	GoFile     string
	StructName string
	TableName  string
	OutFile    string
}

// Field represents a database field
type Field struct {
	Name    string
	GoType  string
	SQLType string
}

// SchemaState represents the current database schema state
type SchemaState struct {
	Meta struct {
		RollbackCount int `json:"rollback_count"`
	} `json:"meta"`
	Tables map[string][]Field `json:"tables"`
}

// Migration represents a migration file
type Migration struct {
	Version  string
	Filename string
	Content  string
}

// CLI flags
var (
	Rollback      bool
	verbose       bool
	dryRun        bool
	targetVersion string
	ConfigPath    string
	strictMode    bool
)

const (
	defaultConfigPath = "config.json"
	versionPrefix     = "version"
)

type SchemaValidator interface {
	Verify(db *sql.DB, tableName string, expectedFields []Field) error
}
