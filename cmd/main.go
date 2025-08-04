package main

import (
	"encoding/gob"
	"log"
	"net/http"

	"bookmaker.ca/internal/cache"
	"bookmaker.ca/internal/db"
	"bookmaker.ca/internal/models"
	"bookmaker.ca/internal/routes"
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
