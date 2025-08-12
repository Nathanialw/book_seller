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

	// Load previous state
	prevState, err := LoadState(config.Paths.StateFile)
	if err != nil {
		log.Printf("Warning: could not load previous state: %v", err)
		prevState = SchemaState{Tables: make(map[string][]Field)}
	}

	// Generate migrations
	forwardMigrations, undoMigrations, newState, err := GenerateMigrations(config, prevState, newVersionPrefix)
	if err != nil {
		return fmt.Errorf("error generating migrations: %w", err)
	}

	if len(forwardMigrations) == 0 && len(undoMigrations) == 0 {
		log.Println("No schema changes detected. No migrations applied.")
		return nil
	}

	// Ensure migration directory exists
	if err := os.MkdirAll(config.Paths.MigrationDir, 0755); err != nil {
		return fmt.Errorf("error creating migration directory: %w", err)
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

			// Verify file exists before execution
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return fmt.Errorf("migration file not found: %s (absolute path: %s)",
					filePath, getAbsolutePath(filePath))
			}

			log.Printf("Applying: %s (absolute path: %s)", migration.Filename, getAbsolutePath(filePath))

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
		// Ensure history directory exists
		if err := os.MkdirAll(config.Paths.HistoryDir, 0755); err != nil {
			return fmt.Errorf("error creating history directory: %w", err)
		}

		// Save history snapshot
		historyFile := filepath.Join(config.Paths.HistoryDir,
			fmt.Sprintf("schema_state_%s.json", prevVersionPrefix))
		if err := SaveState(historyFile, prevState); err != nil {
			return fmt.Errorf("error saving history state: %w", err)
		}

		// Update current state
		if err := SaveState(config.Paths.StateFile, newState); err != nil {
			return fmt.Errorf("error saving current state: %w", err)
		}

		// Update config version
		if err := updateConfigVersion(config.Paths.ConfigFile, newVersion); err != nil {
			return fmt.Errorf("error updating config version: %w", err)
		}
	}

	log.Println("\nMigration complete.")
	log.Printf("Version %s successfully applied.", newVersionPrefix)
	return nil
}

// Helper function to get absolute path
func getAbsolutePath(path string) string {
	absPath, _ := filepath.Abs(path)
	return absPath
}

func pluralize(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		// Category -> categories
		return lower[:len(lower)-1] + "ies"
	}
	// Just add 's' for now
	return lower + "s"
}

func WriteMigrationFiles(migrationDir string, migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("error creating migration directory: %w", err)
	}

	for _, migration := range migrations {
		filePath := filepath.Join(migrationDir, migration.Filename)

		// Write each migration to its own file
		if err := os.WriteFile(filePath, []byte(migration.Content), 0644); err != nil {
			return fmt.Errorf("error writing migration file %s: %w", filePath, err)
		}

		log.Printf("Created migration file: %s", getAbsolutePath(filePath))
	}

	return nil
}

func GenerateMigrations(config *Config, prevState SchemaState, versionPrefix string) ([]Migration, []Migration, SchemaState, error) {
	modelConfigs, err := FindModels(config.Paths.ModelDir, config.Settings.IgnoredStructs, config.Settings.TableNaming)
	if err != nil {
		return nil, nil, SchemaState{}, fmt.Errorf("error finding models: %w", err)
	}

	newState := SchemaState{
		Meta:   prevState.Meta,
		Tables: make(map[string][]Field),
	}

	var allForwardSQL strings.Builder
	var allUndoSQL strings.Builder
	changesDetected := false

	for _, model := range modelConfigs {
		goFilePath := filepath.Join(config.Paths.ModelDir, model.GoFile)
		currentFields, err := ParseStructFields(goFilePath, model.StructName)
		if err != nil {
			return nil, nil, SchemaState{}, fmt.Errorf("error parsing struct fields for %s: %w", model.StructName, err)
		}

		prevFields := prevState.Tables[model.TableName]
		forwardSQL, undoSQL := GenerateSQLStatements(model.TableName, currentFields, prevFields)

		if forwardSQL != "" || undoSQL != "" {
			changesDetected = true

			// Add separator comments between table migrations
			if allForwardSQL.Len() > 0 {
				allForwardSQL.WriteString("\n\n")
				allUndoSQL.WriteString("\n\n")
			}

			allForwardSQL.WriteString(fmt.Sprintf("-- Migration for table: %s\n", model.TableName))
			allForwardSQL.WriteString(forwardSQL)

			allUndoSQL.WriteString(fmt.Sprintf("-- Undo migration for table: %s\n", model.TableName))
			allUndoSQL.WriteString(undoSQL)
		}

		// Update new state
		newState.Tables[model.TableName] = currentFields
	}

	if !changesDetected {
		return nil, nil, SchemaState{}, nil
	}

	// Create single migration files for this version
	forwardMigrations := []Migration{
		{
			Version:  versionPrefix,
			Filename: fmt.Sprintf("%s_%s", versionPrefix, config.Settings.MigrationFile),
			Content:  allForwardSQL.String(),
		},
	}

	undoMigrations := []Migration{
		{
			Version:  versionPrefix,
			Filename: fmt.Sprintf("%s_undo_%s", versionPrefix, config.Settings.MigrationFile),
			Content:  allUndoSQL.String(),
		},
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
				continue
			}

			// Parse field name and type
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			fieldName := parts[0]
			fieldType := parts[1]

			// Parse tags
			var isPrimary bool
			var isNullable bool = true
			var defaultValue string
			var isForeignKey bool
			var references string

			// Automatically set ID fields as primary keys (case-insensitive)
			if strings.EqualFold(fieldName, "id") {
				isPrimary = true
				isNullable = false
			}

			// Automatically set ID fields as primary keys (case-insensitive)
			if strings.EqualFold(fieldName, "id") {
				isPrimary = true
				isNullable = false
			}

			// Detect automatic foreign key based on _ID naming
			if strings.HasSuffix(strings.ToLower(fieldName), "_id") && !strings.EqualFold(fieldName, "id") {
				isForeignKey = true
				baseName := strings.TrimSuffix(strings.ToLower(fieldName), "_id")
				references = fmt.Sprintf("%ss(ID)", baseName) // pluralize by adding "s"
			}

			// Parse tags if present
			if idx := strings.Index(line, "`"); idx != -1 {
				tagSection := line[idx+1:]
				if endIdx := strings.Index(tagSection, "`"); endIdx != -1 {
					tagSection = tagSection[:endIdx]
					tags := strings.Split(tagSection, " ")
					for _, tag := range tags {
						if strings.Contains(tag, "primary") {
							isPrimary = true
							isNullable = false
						}
						if strings.Contains(tag, "default:") {
							defaultValue = strings.TrimPrefix(tag, "default:")
						}
						if strings.Contains(tag, "foreign:") {
							isForeignKey = true
							references = strings.TrimPrefix(tag, "foreign:")
						}
					}
				}
			}

			// Map Go types to SQL types
			sqlType := goTypeToSQL(fieldType)
			if !isNullable && !isPrimary {
				sqlType += " NOT NULL"
			}
			if isPrimary {
				sqlType += " PRIMARY KEY"
			}
			if defaultValue != "" {
				sqlType += " DEFAULT " + defaultValue
			}

			fields = append(fields, Field{
				Name:         fieldName,
				GoType:       fieldType,
				SQLType:      sqlType,
				IsPrimary:    isPrimary,
				IsNullable:   isNullable,
				Default:      defaultValue,
				IsForeignKey: isForeignKey,
				References:   references,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning Go file: %w", err)
	}

	return fields, nil
}

func GenerateSQLStatements(tableName string, currentFields, prevFields []Field) (string, string) {
	var forwardStatements []string
	var undoStatements []string

	// If no previous fields, this is a new table
	if len(prevFields) == 0 {
		// Create table statement
		columns := make([]string, 0, len(currentFields))
		primaryKeys := make([]string, 0)
		foreignKeys := make([]string, 0)

		for _, f := range currentFields {
			columnDef := fmt.Sprintf("%s %s", f.Name, f.SQLType)
			if f.IsPrimary {
				primaryKeys = append(primaryKeys, f.Name)
			}
			columns = append(columns, columnDef)

			if f.IsForeignKey && f.References != "" {
				parts := strings.Split(f.References, "(")
				if len(parts) == 2 {
					refModel := parts[0]
					refColumn := strings.TrimSuffix(parts[1], ")")
					refTable := pluralize(refModel) // pluralize here too
					fkName := fmt.Sprintf("fk_%s_%s_%s", tableName, f.Name, refTable)
					foreignKeys = append(foreignKeys,
						fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)",
							fkName, f.Name, refTable, refColumn))
				}
			}
		}

		// if len(primaryKeys) > 0 {
		// columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
		// }

		columns = append(columns, foreignKeys...)

		forwardStatements = append(forwardStatements,
			fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n);", tableName, strings.Join(columns, ",\n\t")))

		// Undo would drop the table
		undoStatements = append(undoStatements, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", tableName))
	} else {
		// Existing table modification logic
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
				// Add new column
				forwardStatements = append(forwardStatements,
					fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;",
						tableName, cf.Name, cf.SQLType))

				// Add foreign key constraint if needed
				if cf.IsForeignKey && cf.References != "" {
					parts := strings.Split(cf.References, "(")
					if len(parts) == 2 {
						refModel := parts[0]
						refColumn := strings.TrimSuffix(parts[1], ")")
						refTable := pluralize(refModel) // <-- apply pluralize here
						fkName := fmt.Sprintf("fk_%s_%s_%s", tableName, cf.Name, refTable)
						forwardStatements = append(forwardStatements,
							fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s);",
								tableName, fkName, cf.Name, refTable, refColumn))
					}
				}

				undoStatements = append(undoStatements,
					fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s CASCADE;",
						tableName, cf.Name))
			} else if cf.SQLType != pf.SQLType || cf.IsForeignKey != pf.IsForeignKey || cf.References != pf.References {
				// Handle type changes or foreign key changes
				if cf.SQLType != pf.SQLType {
					forwardStatements = append(forwardStatements,
						fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s;",
							tableName, cf.Name, cf.SQLType, cf.Name, cf.SQLType))
				}

				// Handle foreign key changes
				if pf.IsForeignKey {
					// Drop old foreign key if it exists
					fkName := fmt.Sprintf("fk_%s_%s_%s", tableName, cf.Name,
						strings.Split(pf.References, "(")[0])
					forwardStatements = append(forwardStatements,
						fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;",
							tableName, fkName))
				}

				if cf.IsForeignKey && cf.References != "" {
					parts := strings.Split(cf.References, "(")
					if len(parts) == 2 {
						refTable := parts[0]
						refColumn := strings.TrimSuffix(parts[1], ")")
						fkName := fmt.Sprintf("fk_%s_%s_%s", tableName, cf.Name, refTable)
						forwardStatements = append(forwardStatements,
							fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s);",
								tableName, fkName, cf.Name, refTable, refColumn))
					}
				}

				undoStatements = append(undoStatements,
					fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s;",
						tableName, cf.Name, pf.SQLType, cf.Name, pf.SQLType))
			}
		}

		// Removed fields
		for _, pf := range prevFields {
			pfLower := strings.ToLower(pf.Name)
			if _, exists := currentMap[pfLower]; !exists {
				forwardStatements = append(forwardStatements,
					fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s CASCADE;",
						tableName, pf.Name))
				undoStatements = append(undoStatements,
					fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;",
						tableName, pf.Name, pf.SQLType))
			}
		}
	}

	return strings.Join(forwardStatements, "\n"), strings.Join(undoStatements, "\n")
}
