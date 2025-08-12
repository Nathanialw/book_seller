package migrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if not specified
	if config.Settings.VersionPrefixLength == 0 {
		config.Settings.VersionPrefixLength = 5
	}

	return &config, nil
}

func LoadState(stateFile string) (SchemaState, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return SchemaState{}, fmt.Errorf("error reading state file: %w", err)
	}

	var state SchemaState
	if err := json.Unmarshal(data, &state); err != nil {
		return SchemaState{}, fmt.Errorf("error parsing state file: %w", err)
	}

	if state.Tables == nil {
		state.Tables = make(map[string][]Field)
	}

	return state, nil
}

func SaveState(stateFile string, state SchemaState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("error writing state file: %w", err)
	}

	return nil
}

func ParseModelConfigs(configFile string) ([]ModelConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	var configs []ModelConfig
	for _, model := range cfg.Models {
		tableName := model.TableName
		if tableName == "" {
			tableName = structToTableName(model.StructName)
		}

		configs = append(configs, ModelConfig{
			GoFile:     model.GoFile,
			StructName: model.StructName,
			TableName:  tableName,
			OutFile:    model.OutFile,
		})
	}

	return configs, nil
}

func goTypeToSQL(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR"
	case "int", "int32", "int64":
		return "INTEGER"
	case "float32":
		return "FLOAT"
	case "float64":
		return "DOUBLE PRECISION"
	case "bool":
		return "BOOLEAN"
	case "time.Time":
		return "TIMESTAMP"
	default:
		return "TEXT"
	}
}

func getCurrentVersion(configFile string) (int, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return 0, fmt.Errorf("error reading config file: %w", err)
	}

	// Define a minimal struct to extract just the version
	var config struct {
		Version int `json:"version"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return 0, fmt.Errorf("error parsing config file: %w", err)
	}

	return config.Version, nil
}

func updateConfigVersion(configFile string, newVersion int) error {
	// Read the existing config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the JSON into a map for easy manipulation
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Update the version field
	config["version"] = newVersion

	// Marshal back to JSON with indentation
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated config: %w", err)
	}

	// Write the updated config back to file
	if err := os.WriteFile(configFile, updatedData, 0644); err != nil {
		return fmt.Errorf("error writing updated config file: %w", err)
	}

	return nil
}

func determineTargetVersion(currentVersion int) (int, error) {
	if targetVersion == "" {
		// Default to previous version
		target := currentVersion - 1
		if target < 0 {
			target = 0
		}
		return target, nil
	}

	// Parse explicit target version
	target, err := strconv.Atoi(targetVersion)
	if err != nil {
		return 0, fmt.Errorf("invalid target version: %w", err)
	}

	if target < 0 {
		return 0, errors.New("target version cannot be negative")
	}

	return target, nil
}

func structToTableName(structName string) string {
	// Convert camelCase to snake_case and pluralize
	var result []rune
	for i, r := range structName {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	tableName := string(result)

	// Basic pluralization
	if !strings.HasSuffix(tableName, "s") {
		tableName += "s"
	}
	return tableName
}
