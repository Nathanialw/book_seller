package migrations

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func Init() {
	flag.BoolVar(&Rollback, "r", false, "Rollback mode")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&dryRun, "d", false, "Dry run (don't execute changes)")
	flag.StringVar(&targetVersion, "t", "", "Target version for rollback")
	flag.StringVar(&ConfigPath, "c", defaultConfigPath, "Path to config file")
	flag.BoolVar(&strictMode, "strict", false, "Abort if models don't match target version")
}

// Then fix the initStateFile function:
func InitStateFile(path string) error {
	initialState := SchemaState{
		Meta: struct {
			RollbackCount int `json:"rollback_count"`
		}{
			RollbackCount: 0,
		},
		Tables: make(map[string][]Field),
	}

	data, err := json.MarshalIndent(initialState, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling initial state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing initial state file: %w", err)
	}

	return nil
}

func ExecuteSQL(config *Config, sqlFile, direction string) error {
	if dryRun {
		log.Printf("[DRY RUN] Would execute %s migration: %s", direction, sqlFile)
		return nil
	}

	cmd := exec.Command("psql",
		"-h", config.Database.Host,
		"-p", config.Database.Port,
		"-U", config.Database.User,
		"-d", config.Database.Name,
		"-f", sqlFile)

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", config.Database.Password),
		fmt.Sprintf("PGSSLMODE=%s", config.Database.SSLMode))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing SQL: %w\nOutput: %s", err, string(output))
	}

	if verbose {
		log.Printf("SQL execution output:\n%s", string(output))
	}

	return nil
}

func EnsureDirs(config *Config) error {
	dirs := []string{
		config.Paths.HistoryDir,
		config.Paths.MigrationDir,
		config.Paths.ModelDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}
	return nil
}

func VerifySchemaOnStart(config *Config) error {
	db, err := sql.Open("postgres", BuildDSN(config))
	if err != nil {
		return err
	}
	defer db.Close()

	validator := PostgresValidator{}
	modelConfigs, err := ParseModelConfigs(config.Paths.ConfigFile)
	if err != nil {
		return err
	}

	currentState, err := LoadState(config.Paths.StateFile)
	if err != nil {
		return err
	}

	for _, model := range modelConfigs {
		expectedFields := currentState.Tables[model.TableName]
		if err := validator.Verify(db, model.TableName, expectedFields); err != nil {
			return fmt.Errorf("table %s: %w", model.TableName, err)
		}
	}

	return nil
}

func BuildDSN(config *Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode)
}
