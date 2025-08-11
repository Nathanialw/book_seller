// validations.go
package migrations

import (
	"database/sql"
	"fmt"
	"strings"
)

type PostgresValidator struct{}

func (v PostgresValidator) Verify(db *sql.DB, tableName string, expectedFields []Field) error {
	query := `
        SELECT column_name, data_type 
        FROM information_schema.columns 
        WHERE table_name = $1
    `
	rows, err := db.Query(query, tableName)
	if err != nil {
		return fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

	dbFields := make(map[string]string)
	for rows.Next() {
		var name, dtype string
		if err := rows.Scan(&name, &dtype); err != nil {
			return err
		}
		dbFields[strings.ToLower(name)] = strings.ToUpper(dtype)
	}

	for _, expected := range expectedFields {
		sqlType := goTypeToSQL(expected.GoType)
		dbType, exists := dbFields[strings.ToLower(expected.Name)]

		if !exists {
			return fmt.Errorf("missing column: %s (expected %s)", expected.Name, sqlType)
		}

		if !TypeMatches(dbType, sqlType) {
			return fmt.Errorf("type mismatch for %s: DB has %s, model wants %s",
				expected.Name, dbType, sqlType)
		}
	}

	return nil
}

func TypeMatches(dbType string, expectedType string) bool {
	// Normalize types
	dbType = strings.ToUpper(strings.TrimSpace(dbType))
	expectedType = strings.ToUpper(strings.TrimSpace(expectedType))

	// Remove length constraints for comparison
	dbBaseType := strings.Split(dbType, "(")[0]
	expectedBaseType := strings.Split(expectedType, "(")[0]

	// Simple type compatibility matrix
	compatibleTypes := map[string][]string{
		"CHAR":      {"CHAR", "VARCHAR", "TEXT", "CHARACTER", "CHARACTER VARYING"},
		"VARCHAR":   {"VARCHAR", "CHAR", "TEXT", "CHARACTER", "CHARACTER VARYING"},
		"TEXT":      {"TEXT", "VARCHAR", "CHAR", "CHARACTER", "CHARACTER VARYING"},
		"INTEGER":   {"INT", "INTEGER", "BIGINT", "SMALLINT"},
		"TIMESTAMP": {"TIMESTAMP", "TIMESTAMPTZ", "DATE"},
		"BOOLEAN":   {"BOOL", "BOOLEAN"},
	}

	// Exact match
	if dbType == expectedType {
		return true
	}

	// Base type match
	if dbBaseType == expectedBaseType {
		return true
	}

	// Check compatible groups
	for _, group := range compatibleTypes {
		if Contains(group, dbBaseType) && Contains(group, expectedBaseType) {
			return true
		}
	}

	return false
}

func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
