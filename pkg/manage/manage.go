package manage

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/nathanialw/ecommerce/internal/cache"
	"github.com/nathanialw/ecommerce/internal/db"
	"github.com/nathanialw/ecommerce/pkg/models"
	"github.com/nathanialw/ecommerce/pkg/routes"
)

func Run() {
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
