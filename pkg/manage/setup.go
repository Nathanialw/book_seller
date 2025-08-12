package manage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/nathanialw/ecommerce/internal/migrations"
)

func Setup(config *migrations.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := setupDatabase(ctx, *config); err != nil {
		return fmt.Errorf("database setup failed: %w", err)
	}

	// After database is created, execute schema files
	if err := executeSchemaFiles(ctx, *config); err != nil {
		return fmt.Errorf("schema execution failed: %w", err)
	}

	log.Println("✅ All setup completed successfully")
	return nil
}

func executeSchemaFiles(ctx context.Context, config migrations.Config) error {
	appDB, err := getPostgresConnection(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect to application database: %w", err)
	}
	defer appDB.Close()

	//TODO:
	//save schemaDir in a config file
	// Get the path to the schema directory relative to this source file
	_, filename, _, _ := runtime.Caller(0) // Gets path to current file
	dir := filepath.Dir(filename)
	schemaDir := filepath.Join(dir, "schemas")

	files, err := os.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	// Sort files by name to ensure proper order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			content, err := os.ReadFile(filepath.Join(schemaDir, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to read schema file %s: %w", file.Name(), err)
			}

			log.Printf("🔧 Executing schema file: %s", file.Name())
			if _, err := appDB.ExecContext(ctx, string(content)); err != nil {
				return fmt.Errorf("failed to execute schema file %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

func getPostgresConnection(ctx context.Context, config migrations.Config) (*sql.DB, error) {
	port, _ := strconv.Atoi(config.Database.Port) // default

	attempts := []string{
		// Try with password and all parameters first
		fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			port,
			config.Database.Name,
			config.Database.SSLMode),

		// Fallback to socket connection if host is local
		fmt.Sprintf("user=%s host=/var/run/postgresql dbname=%s sslmode=disable",
			config.Database.User,
			config.Database.Name),
	}

	for _, dsn := range attempts {
		db, err := sql.Open("postgres", dsn) // Note: "postgres" driver, not "eccomerce"
		if err != nil {
			log.Printf("Connection attempt failed with DSN %q: %v", dsn, err)
			continue
		}

		// Verify connection works
		if err := db.PingContext(ctx); err == nil {
			return db, nil
		}
		log.Printf("Ping failed with DSN %q: %v", dsn, err)
		db.Close()
	}
	return nil, fmt.Errorf("could not connect to PostgreSQL after %d attempts", len(attempts))
}

func setupDatabase(ctx context.Context, config migrations.Config) error {
	log.Println("==> Verifying PostgreSQL connection...")

	adminDB, err := getPostgresConnection(ctx, config)
	if err != nil {
		return fmt.Errorf("postgresql verification failed: %w", err)
	}
	defer adminDB.Close()
	log.Println("✅ Successfully connected to PostgreSQL")

	log.Println("==> Setting up database and user...")

	if err := createDatabase(ctx, adminDB, config.Database.Name); err != nil {
		return fmt.Errorf("database creation failed: %w", err)
	}

	if err := createUser(ctx, adminDB, config.Database.User, config.Database.Password); err != nil {
		return fmt.Errorf("user creation failed: %w", err)
	}

	if _, err := adminDB.ExecContext(ctx,
		fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s",
			config.Database.Name, config.Database.User)); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	return nil
}

func createDatabase(ctx context.Context, db *sql.DB, name string) error {
	var exists bool
	err := db.QueryRowContext(ctx,
		"SELECT 1 FROM pg_database WHERE datname = $1", name).Scan(&exists)

	if err == sql.ErrNoRows || !exists {
		log.Printf("🔧 Creating database '%s'...", name)
		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", name))
		return err
	}
	if err != nil {
		return err
	}
	log.Printf("✅ Database '%s' already exists", name)
	return nil
}

func createUser(ctx context.Context, db *sql.DB, username, password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	var exists bool
	err := db.QueryRowContext(ctx,
		"SELECT 1 FROM pg_roles WHERE rolname = $1", username).Scan(&exists)

	if err == sql.ErrNoRows || !exists {
		log.Printf("🔧 Creating user '%s'...", username)
		_, err = db.ExecContext(ctx,
			fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", username, password))
		return err
	}
	if err != nil {
		return err
	}
	log.Printf("✅ User '%s' already exists", username)
	return nil
}
