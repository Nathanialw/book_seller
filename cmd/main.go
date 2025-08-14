package main

import (
	"github.com/nathanialw/ecommerce/pkg/manage"
)

func main() {
	manage.Run()
	// config, err := manage.Init()
	// if err != nil {
	// 	log.Fatalf("Initialization failed: %v", err)
	// }

	// if err := manage.Setup(config); err != nil {
	// 	log.Fatalf("Setup failed: %v", err)
	// }

	// migrations.Migrate(config)

}
