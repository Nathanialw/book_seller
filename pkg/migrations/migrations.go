package migrations

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nathanialw/ecommerce/internal/migrations"
)

func Migrate() {

	fmt.Println("commade:", os.Args[1])
	fmt.Println("parameter:", os.Args[2])

	migrations.Init()
	flag.Parse()
	// Load configuration
	config, err := migrations.LoadConfig(migrations.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure directories exist
	if err := migrations.EnsureDirs(config); err != nil {
		log.Fatalf("Directory setup failed: %v", err)
	}

	// Initialize state file if it doesn't exist
	if _, err := os.Stat(config.Paths.StateFile); os.IsNotExist(err) {
		if err := migrations.InitStateFile(config.Paths.StateFile); err != nil {
			log.Fatalf("Failed to initialize state file: %v", err)
		}
	}

	if err := migrations.VerifySchemaOnStart(config); err != nil {
		log.Fatalf("Schema verification failed: %v", err)
	}

	if migrations.Rollback {
		if err := migrations.HandleRollback(config); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
	} else {
		if err := migrations.HandleMigration(config); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
}
