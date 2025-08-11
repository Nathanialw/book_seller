package migrations

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func HandleMigration(config *Config) error {
	currentVersion, err := getCurrentVersion(config.Paths.ConfigFile)
	if err != nil {
		return fmt.Errorf("error getting current version: %w", err)
	}

	newVersion := currentVersion + 1
	prevVersionPrefix := fmt.Sprintf("%0*d", config.Settings.VersionPrefixLength, currentVersion)
	newVersionPrefix := fmt.Sprintf("%0*d", config.Settings.VersionPrefixLength, newVersion)

	// Load model configurations
	modelConfigs, err := ParseModelConfigs(config.Paths.ConfigFile)
	if err != nil {
		return fmt.Errorf("error parsing model configs: %w", err)
	}

	// Load previous state
	prevState, err := LoadState(config.Paths.StateFile)
	if err != nil {
		// return fmt.Errorf("error loading previous state: %w", err)
	}

	// Generate migrations
	forwardMigrations, undoMigrations, newState, err := GenerateMigrations(config, modelConfigs, prevState, newVersionPrefix, prevVersionPrefix)
	if err != nil {
		return fmt.Errorf("error generating migrations: %w", err)
	}

	if len(forwardMigrations) == 0 && len(undoMigrations) == 0 {
		log.Println("No schema changes detected. No migrations applied.")
		return nil
	}

	// Write migration files
	if err := WriteMigrationFiles(config.Paths.MigrationDir, forwardMigrations); err != nil {
		return fmt.Errorf("error writing migration files: %w", err)
	}

	if err := WriteMigrationFiles(config.Paths.MigrationDir, undoMigrations); err != nil {
		return fmt.Errorf("error writing undo migration files: %w", err)
	}

	// Execute migrations
	if !dryRun {
		for _, migration := range forwardMigrations {
			filePath := filepath.Join(config.Paths.MigrationDir, migration.Filename)
			log.Printf("Applying: %s", migration.Filename)

			if err := ExecuteSQL(config, filePath, "up"); err != nil {
				return fmt.Errorf("error executing migration: %w", err)
			}
		}
	} else {
		for _, migration := range forwardMigrations {
			log.Printf("[DRY RUN] Would apply migration: %s", migration.Filename)
		}
	}

	// Update state
	if !dryRun {
		// Save history snapshot
		historyFile := filepath.Join(config.Paths.HistoryDir, fmt.Sprintf("schema_state_%s.json", prevVersionPrefix))
		if err := SaveState(historyFile, prevState); err != nil {
			return fmt.Errorf("error saving history state: %w", err)
		}
		// Update current state
		if err := SaveState(config.Paths.StateFile, newState); err != nil {
			return fmt.Errorf("error saving current state: %w", err)
		}
		println(newVersion)

		// Update config version
		if err := updateConfigVersion(config.Paths.ConfigFile, newVersion); err != nil {
			return fmt.Errorf("error updating config version: %w", err)
		}
	}

	// Print summary
	log.Println("\nMigration complete.")
	log.Printf("Version %s successfully applied.", newVersionPrefix)
	log.Println("Created:")
	for _, m := range forwardMigrations {
		log.Printf("  - %s", filepath.Join(config.Paths.MigrationDir, m.Filename))
	}
	for _, m := range undoMigrations {
		log.Printf("  - %s", filepath.Join(config.Paths.MigrationDir, m.Filename))
	}

	return nil
}

func GenerateMigrations(config *Config, modelConfigs []ModelConfig, prevState SchemaState, versionPrefix, prevVersionPrefix string) ([]Migration, []Migration, SchemaState, error) {
	var forwardMigrations []Migration
	var undoMigrations []Migration
	newState := SchemaState{
		Meta:   prevState.Meta,
		Tables: make(map[string][]Field),
	}

	changesDetected := false

	for _, model := range modelConfigs {
		goFilePath := filepath.Join(config.Paths.ModelDir, model.GoFile)
		currentFields, err := ParseStructFields(goFilePath, model.StructName)

		if err != nil {
			return nil, nil, SchemaState{}, fmt.Errorf("error parsing struct fields: %w", err)
		}

		prevFields := prevState.Tables[model.TableName]

		// Generate SQL statements
		forwardSQL, undoSQL := GenerateSQLStatements(model.TableName, currentFields, prevFields)

		if forwardSQL != "" || undoSQL != "" {
			changesDetected = true

			// Create forward migration
			if forwardSQL != "" {
				forwardMigrations = append(forwardMigrations, Migration{
					Version:  versionPrefix,
					Filename: fmt.Sprintf("%s_%s", versionPrefix, model.OutFile),
					Content:  fmt.Sprintf("%s", forwardSQL),
				})
			}

			// Create undo migration
			if undoSQL != "" {
				undoFilename := fmt.Sprintf("%s_%s_undo.sql", versionPrefix, strings.TrimSuffix(model.OutFile, ".sql"))
				undoMigrations = append(undoMigrations, Migration{
					Version:  versionPrefix,
					Filename: undoFilename,
					Content:  fmt.Sprintf("%s", undoSQL),
				})
			}
		}

		// Update new state
		newState.Tables[model.TableName] = currentFields
	}

	if !changesDetected {
		return nil, nil, SchemaState{}, nil
	}
	for _, v := range forwardMigrations {
		println("forwardStatements: ", v.Content)
	}

	return forwardMigrations, undoMigrations, newState, nil
}

func ParseStructFields(goFilePath, structName string) ([]Field, error) {
	data, err := os.ReadFile(goFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading Go file: %w", err)
	}

	var fields []Field
	inStruct := false
	structPattern := regexp.MustCompile(`type\s+` + structName + `\s+struct\s*\{`)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		// Detect struct start
		if structPattern.MatchString(line) {
			inStruct = true
			continue
		}

		// Detect struct end
		if inStruct && strings.HasPrefix(line, "}") {
			break
		}

		if inStruct {
			// Skip embedded structs
			if strings.Contains(line, ".") {
				println(line, " is an embedded struct")
				continue
			}

			// Parse field name and type
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			fieldName := parts[0]
			fieldType := parts[1]

			// Strip tags
			if idx := strings.Index(fieldType, "`"); idx != -1 {
				fieldType = fieldType[:idx]
			}
			if idx := strings.Index(fieldType, "//"); idx != -1 {
				fieldType = fieldType[:idx]
			}

			// Map Go types to SQL types
			sqlType := goTypeToSQL(fieldType)

			fields = append(fields, Field{
				Name:    fieldName,
				GoType:  fieldType,
				SQLType: sqlType,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning Go file: %w", err)
	}

	return fields, nil
}

func WriteMigrationFiles(migrationDir string, migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	var filename string
	for _, migration := range migrations {
		filename = migration.Filename
	}

	// Determine the combined filename based on the first migration's version
	version := migrations[0].Version
	combinedFilename := fmt.Sprintf("%s", filename)
	combinedPath := filepath.Join(migrationDir, combinedFilename)

	var combinedContent bytes.Buffer
	combinedContent.WriteString(fmt.Sprintf("-- Combined Migrations Version: %s\n", version))
	combinedContent.WriteString("-- This file contains all migrations for this version\n\n")

	// Append all migrations to the combined file
	for _, migration := range migrations {
		combinedContent.WriteString(migration.Content)
		combinedContent.WriteString("\n")
	}

	// Write the combined file
	if err := os.WriteFile(combinedPath, combinedContent.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing combined migration file %s: %w", combinedPath, err)
	}

	return nil
}

func GenerateSQLStatements(tableName string, currentFields, prevFields []Field) (string, string) {
	var forwardStatements []string
	var undoStatements []string

	currentMap := make(map[string]Field)
	prevMap := make(map[string]Field)

	for _, f := range currentFields {
		currentMap[strings.ToLower(f.Name)] = f
	}

	for _, f := range prevFields {
		prevMap[strings.ToLower(f.Name)] = f
	}

	// Added or changed fields
	for _, cf := range currentFields {
		cfLower := strings.ToLower(cf.Name)
		pf, exists := prevMap[cfLower]

		if !exists {
			// Field added
			forwardStatements = append(forwardStatements, fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;", tableName, cf.Name, cf.SQLType))
			undoStatements = append(undoStatements, fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s CASCADE;", tableName, cf.Name))
		} else if cf.SQLType != pf.SQLType {
			// Field type changed
			forwardStatements = append(forwardStatements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s;", tableName, cf.Name, cf.SQLType, cf.Name, cf.SQLType))
			undoStatements = append(undoStatements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s;", tableName, cf.Name, pf.SQLType, cf.Name, pf.SQLType))
		}
	}

	// Removed fields
	for _, pf := range prevFields {
		pfLower := strings.ToLower(pf.Name)
		if _, exists := currentMap[pfLower]; !exists {
			// Field removed
			forwardStatements = append(forwardStatements, fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s CASCADE;", tableName, pf.Name))
			undoStatements = append(undoStatements, fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;", tableName, pf.Name, pf.SQLType))
		}
	}

	return strings.Join(forwardStatements, "\n"), strings.Join(undoStatements, "\n")
}
