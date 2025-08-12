package main

import (
	"github.com/nathanialw/ecommerce/pkg/manage"
	"github.com/nathanialw/ecommerce/pkg/migrations"
)

func main() {
	// manage.Run()
	config, _ := manage.Init()
	manage.Setup(config)

	migrations.Migrate(config)

}
