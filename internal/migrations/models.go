package migrations

import (
	"bufio"
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

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
		VersionPrefixLength int      `json:"version_prefix_length"`
		MigrationFile       string   `json:"migration_file"`
		TableNaming         string   `json:"table_naming"`
		IgnoredStructs      []string `json:"ignored_structs"`
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
	Name         string
	GoType       string
	SQLType      string
	IsPrimary    bool
	IsNullable   bool
	Default      string
	IsForeignKey bool
	References   string // Format: "table(column)"
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

func FindModels(modelDir string, ignoredStructs []string, namingConvention string) ([]ModelConfig, error) {
	var configs []ModelConfig

	err := filepath.Walk(modelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			structs, err := findStructsInFile(path, ignoredStructs)
			if err != nil {
				return err
			}

			for _, s := range structs {
				configs = append(configs, ModelConfig{
					StructName: s,
					TableName:  generateTableName(s, namingConvention),
					GoFile:     info.Name(),
					OutFile:    "migrations.sql", // Default value
				})
			}
		}
		return nil
	})

	return configs, err
}

func findStructsInFile(filePath string, ignoredStructs []string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var structs []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	structPattern := regexp.MustCompile(`type\s+([A-Z][a-zA-Z0-9]*)\s+struct`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := structPattern.FindStringSubmatch(line); matches != nil {
			structName := matches[1]
			if !contains(ignoredStructs, structName) {
				structs = append(structs, structName)
			}
		}
	}

	return structs, scanner.Err()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateTableName converts a struct name to a database table name based on naming convention
func generateTableName(structName, namingConvention string) string {
	switch namingConvention {
	case "snake_case":
		return toSnakeCase(structName)
	case "snake_case_plural":
		name := toSnakeCase(structName)
		return pluralizeName(name)
	case "camelCase":
		return string(unicode.ToLower(rune(structName[0]))) + structName[1:]
	default: // default to snake_case_plural
		name := toSnakeCase(structName)
		return pluralizeName(name)
	}
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// pluralizeName handles basic pluralization of table names
func pluralizeName(name string) string {
	if strings.HasSuffix(name, "s") || strings.HasSuffix(name, "x") ||
		strings.HasSuffix(name, "z") || strings.HasSuffix(name, "ch") ||
		strings.HasSuffix(name, "sh") {
		return name + "es"
	} else if strings.HasSuffix(name, "y") {
		return strings.TrimSuffix(name, "y") + "ies"
	}
	return name + "s"
}
