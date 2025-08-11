package migrations

import (
	"github.com/nathanialw/ecommerce/internal/migrations"
)

func Migrate(config *migrations.Config) {
	migrations.Migrate(config)

}
