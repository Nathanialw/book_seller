package manage

import (
	"encoding/gob"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nathanialw/ecommerce/internal/cache"
	"github.com/nathanialw/ecommerce/internal/db"
	"github.com/nathanialw/ecommerce/internal/migrations"
	"github.com/nathanialw/ecommerce/pkg/models"
	"github.com/nathanialw/ecommerce/pkg/routes"
)

func Init() (*migrations.Config, error) {
	migrations.Init()
	config, err := migrations.LoadConfig(migrations.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return config, err
}

func Run() (*mux.Router, *pgxpool.Pool) {
	gob.Register([]models.CartItem{})

	db := db.InitDB()
	if err := cache.LoadCache(); err != nil {
		log.Fatalf("Failed to load genres: %v", err)
	}

	r := routes.SetupRoutes()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Starting server on :6600")
	http.ListenAndServe(":6600", r)

	return r, db
}
