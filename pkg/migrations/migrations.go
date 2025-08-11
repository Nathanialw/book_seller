package migrations

import (
	"log"

	"github.com/nathanialw/ecommerce/internal/migrations"
)

func Init() (*migrations.Config, error) {
	migrations.Init()
	config, err := migrations.LoadConfig(migrations.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return config, err
}

func Migrate(config *migrations.Config) {
	migrations.Migrate(config)

}
