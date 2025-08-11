package manage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	log.Println("✅ All setup completed successfully")
	return nil
}

func getPostgresConnection(ctx context.Context, config migrations.Config) (*sql.DB, error) {
	attempts := []string{
		"user=postgres host=/var/run/postgresql sslmode=disable",
		fmt.Sprintf("user=postgres host=%s port=%d sslmode=disable",
			config.Database.Host, config.Database.Port),
	}

	for _, dsn := range attempts {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			continue
		}

		if err := db.PingContext(ctx); err == nil {
			return db, nil
		}
		db.Close()
	}
	return nil, fmt.Errorf("could not connect to PostgreSQL")
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
