package main

import (
	"github.com/nathanialw/ecommerce/pkg/manage"
)

func main() {
	// manage.Run()
	config, _ := manage.Init()
	manage.Setup(config)

}
