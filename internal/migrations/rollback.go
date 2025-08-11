package migrations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func HandleRollback(config *Config) error {
	println(config.Paths.ConfigFile)

	currentVersion, err := getCurrentVersion(config.Paths.ConfigFile)

	if err != nil {
		return fmt.Errorf("error getting current version: %w", err)
	}

	targetVer, err := determineTargetVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("error determining target version: %w", err)
	}

	if targetVer >= currentVersion {
		return fmt.Errorf("current version is %d, cannot roll forward to %d", currentVersion, targetVer)
	}

	if config.StrictMode {
		if err := verifyModelsMatchTarget(config, targetVer); err != nil {
			return fmt.Errorf("strict mode: %w", err)
		}
	}

	log.Printf("Rolling back from version %d to %d", currentVersion, targetVer)

	// Load current state
	currentState, err := LoadState(config.Paths.StateFile)
	if err != nil {
		return fmt.Errorf("error loading current state: %w", err)
	}

	// Increment rollback count
	currentState.Meta.RollbackCount++

	// Process each version to roll back
	for v := currentVersion; v > targetVer; v-- {
		versionPrefix := fmt.Sprintf("%0*d", config.Settings.VersionPrefixLength, v)

		undoFiles, err := findUndoFiles(config.Paths.MigrationDir, versionPrefix)
		if err != nil {
			return fmt.Errorf("error finding undo files: %w", err)
		}

		if len(undoFiles) == 0 {
			return fmt.Errorf("no undo files found for version %s", versionPrefix)
		}

		for _, file := range undoFiles {
			log.Printf("Executing rollback: %s", file)

			if !dryRun {
				if err := ExecuteSQL(config, file, "down"); err != nil {
					return fmt.Errorf("error executing rollback script: %w", err)
				}
			} else {
				log.Printf("[DRY RUN] Would execute rollback: %s", file)
			}

			// Archive files
			if !dryRun {
				if err := archiveMigrationFiles(config, v, currentState.Meta.RollbackCount); err != nil {
					return fmt.Errorf("error archiving migration files: %w", err)
				}
			}
		}
	}

	// Update state after successful rollback
	if !dryRun {
		if targetVer == 0 {
			// Reset to empty state but keep rollback count
			currentState.Tables = make(map[string][]Field)
		} else {
			// Load historical state
			historyFile := filepath.Join(config.Paths.HistoryDir, fmt.Sprintf("schema_state_%0*d.json",
				config.Settings.VersionPrefixLength, targetVer))

			historicalState, err := LoadState(historyFile)
			if err != nil {
				return fmt.Errorf("error loading historical state: %w", err)
			}

			// Preserve the rollback count
			historicalState.Meta.RollbackCount = currentState.Meta.RollbackCount
			currentState = historicalState
		}

		// Save state
		if err := SaveState(config.Paths.StateFile, currentState); err != nil {
			return fmt.Errorf("error saving state: %w", err)
		}

		// Update config version
		if err := updateConfigVersion(config.Paths.ConfigFile, targetVer); err != nil {
			return fmt.Errorf("error updating config version: %w", err)
		}

		// Verify rollback
		if err := VerifyRollback(config, targetVer); err != nil {
			return fmt.Errorf("rollback verification failed: %w", err)
		}
	}

	log.Printf("Successfully rolled back to version %d", targetVer)
	log.Printf("Rollback count: %d", currentState.Meta.RollbackCount)
	return nil
}

func VerifyRollback(config *Config, targetVersion int) error {
	// Verify config version
	currentVersion, err := getCurrentVersion(config.Paths.ConfigFile)
	if err != nil {
		return fmt.Errorf("error getting current version: %w", err)
	}

	if currentVersion != targetVersion {
		return fmt.Errorf("config version mismatch after rollback (expected %d, got %d)",
			targetVersion, currentVersion)
	}

	// Verify state file
	if targetVersion == 0 {
		state, err := LoadState(config.Paths.StateFile)
		if err != nil {
			return fmt.Errorf("error loading state file: %w", err)
		}

		if len(state.Tables) > 0 {
			return errors.New("state file not empty after rollback to version 0")
		}
	} else {
		historyFile := filepath.Join(config.Paths.HistoryDir,
			fmt.Sprintf("schema_state_%0*d.json", config.Settings.VersionPrefixLength, targetVersion))

		currentState, err := LoadState(config.Paths.StateFile)
		if err != nil {
			return fmt.Errorf("error loading current state: %w", err)
		}

		historicalState, err := LoadState(historyFile)
		if err != nil {
			return fmt.Errorf("error loading historical state: %w", err)
		}

		// Compare states (ignore rollback count)
		currentState.Meta.RollbackCount = 0
		historicalState.Meta.RollbackCount = 0

		currentJSON, _ := json.Marshal(currentState)
		historicalJSON, _ := json.Marshal(historicalState)

		if !bytes.Equal(currentJSON, historicalJSON) {
			return errors.New("state file does not match history for target version")
		}
	}

	return nil
}

func archiveMigrationFiles(config *Config, version, rollbackCount int) error {
	versionPrefix := fmt.Sprintf("%0*d", config.Settings.VersionPrefixLength, version)

	// Create archive directory if it doesn't exist
	archiveDir := filepath.Join(config.Paths.MigrationDir, "archived")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("error creating archive directory: %w", err)
	}

	// Find all migration files for this version
	pattern := filepath.Join(config.Paths.MigrationDir, fmt.Sprintf("%s_*", versionPrefix))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("error finding migration files: %w", err)
	}

	// Move files to archive with rollback count prefix
	for _, file := range files {
		baseName := filepath.Base(file)
		newPath := filepath.Join(archiveDir, fmt.Sprintf("%d_%s", rollbackCount, baseName))

		if err := os.Rename(file, newPath); err != nil {
			return fmt.Errorf("error archiving file %s: %w", file, err)
		}
	}

	// Archive state file
	stateFile := filepath.Join(config.Paths.HistoryDir, fmt.Sprintf("schema_state_%s.json", versionPrefix))
	if _, err := os.Stat(stateFile); err == nil {
		newPath := filepath.Join(archiveDir, fmt.Sprintf("%d_schema_state_%s.json", rollbackCount, versionPrefix))
		if err := os.Rename(stateFile, newPath); err != nil {
			return fmt.Errorf("error archiving state file: %w", err)
		}
	}

	return nil
}

func findUndoFiles(migrationDir, versionPrefix string) ([]string, error) {
	pattern := filepath.Join(migrationDir, fmt.Sprintf("%s_*_undo.sql", versionPrefix))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error finding undo files: %w", err)
	}
	return matches, nil
}

func verifyModelsMatchTarget(config *Config, targetVersion int) error {
	// 1. Load model configs
	modelConfigs, err := ParseModelConfigs(config.Paths.ConfigFile)
	if err != nil {
		return err
	}

	// 2. Load historical state for target version
	historyFile := filepath.Join(config.Paths.HistoryDir,
		fmt.Sprintf("schema_state_%0*d.json",
			config.Settings.VersionPrefixLength, targetVersion))

	historicalState, err := LoadState(historyFile)
	if err != nil {
		return err
	}

	// 3. Verify each model
	for _, model := range modelConfigs {
		currentFields, err := ParseStructFields(
			filepath.Join(config.Paths.ModelDir, model.GoFile),
			model.StructName)
		if err != nil {
			return err
		}

		historicalFields := historicalState.Tables[model.TableName]
		if !fieldsMatch(currentFields, historicalFields) {
			return fmt.Errorf("model %s doesn't match target version %d",
				model.StructName, targetVersion)
		}
	}

	return nil
}

func fieldsMatch(a, b []Field) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Name != b[i].Name || a[i].GoType != b[i].GoType {
			return false
		}
	}

	return true
}
