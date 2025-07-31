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
	r.HandleFunc("/book/{id}", handlers.BookDetailHandler).Methods("GET")
	r.HandleFunc("/about", handlers.AboutHandler).Methods("GET")
	r.HandleFunc("/videos", handlers.VideosHandler).Methods("GET")
	r.HandleFunc("/search", handlers.SearchHandler).Methods("GET")
	r.HandleFunc("/admin/login", handlers.AdminLoginGet).Methods("GET")
	r.HandleFunc("/admin", handlers.AdminHandler).Methods("GET")
	r.HandleFunc("/admin/add-book", handlers.AddBookForm).Methods("GET")

	r.HandleFunc("/admin/AdminLogin", handlers.AdminLoginHandler).Methods("POST")
	r.HandleFunc("/admin/add-book", handlers.AddBookSubmit).Methods("POST")
	r.HandleFunc("/checkout", handlers.CreateCheckoutSession).Methods("POST")
	r.HandleFunc("/SuccessHandler", handlers.SuccessHandler).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	return r
}
