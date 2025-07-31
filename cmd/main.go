package main

import (
	"log"
	"net/http"

	"bookmaker.ca/internal/routes"
)

func main() {
	r := routes.SetupRoutes()
	log.Println("Starting server on :6600")
	http.ListenAndServe(":6600", r)
}
