package main

import (
	"log"
	"net/http"

	"bookmaker.ca/internal/db"
	"bookmaker.ca/internal/routes"
)

func main() {
	r := routes.SetupRoutes()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	db.InitDB()

	log.Println("Starting server on :6600")
	http.ListenAndServe(":6600", r)
}
