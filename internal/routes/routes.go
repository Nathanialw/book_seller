package routes

import (
	"net/http"

	"bookmaker.ca/internal/handlers"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/books", handlers.BookListHandler).Methods("GET")
	// r.HandleFunc("/book/{id}", handlers.BookDetailHandler).Methods("GET")
	r.HandleFunc("/checkout", handlers.CreateCheckoutSession).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	return r
}
