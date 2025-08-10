package main

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/Nathanialw/ecommerce/internal/cache"
	"github.com/Nathanialw/ecommerce/internal/db"
	"github.com/Nathanialw/ecommerce/internal/models"
	"github.com/Nathanialw/ecommerce/internal/routes"
)

func main() {
	gob.Register([]models.CartItem{})

	db.InitDB()
	if err := cache.LoadCache(); err != nil {
		log.Fatalf("Failed to load genres: %v", err)
	}

	r := routes.SetupRoutes()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Starting server on :6600")
	http.ListenAndServe(":6600", r)
}
