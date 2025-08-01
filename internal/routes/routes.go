package routes

import (
	"net/http"

	"bookmaker.ca/internal/handlers"

	"github.com/gorilla/mux"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		// TODO:
		// validate session value
		next(w, r)
	}
}

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/about", handlers.AboutHandler).Methods("GET")
	r.HandleFunc("/videos", handlers.VideosHandler).Methods("GET")
	r.HandleFunc("/blogs", handlers.BlogsHandler).Methods("GET")

	r.HandleFunc("/books", handlers.BookListHandler).Methods("GET")
	r.HandleFunc("/book/{id}", handlers.BookDetailHandler).Methods("GET")
	r.HandleFunc("/search", handlers.SearchHandler).Methods("GET")

	// Payment
	r.HandleFunc("/checkout", handlers.CreateCheckoutSession).Methods("POST")
	r.HandleFunc("/SuccessHandler", handlers.SuccessHandler).Methods("POST")

	// Admin
	admin := r.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/login", handlers.AdminLoginHandler).Methods("GET")
	admin.HandleFunc("/AdminLogin", handlers.AdminLoginValidateHandler).Methods("POST")
	admin.HandleFunc("/logout", RequireAuth(handlers.AdminLogoutHandler)).Methods("GET")
	admin.HandleFunc("/blogs", RequireAuth(handlers.AdminBlogHandler)).Methods("GET")
	admin.HandleFunc("/videos", RequireAuth(handlers.AdminVideosHandler)).Methods("GET")
	admin.HandleFunc("", RequireAuth(handlers.AdminHandler)).Methods("GET")
	admin.HandleFunc("/add-book", RequireAuth(handlers.AddBookForm)).Methods("GET")
	admin.HandleFunc("/add-book", RequireAuth(handlers.AddBookHandler)).Methods("POST")
	admin.HandleFunc("/update-book", RequireAuth(handlers.UpdateBookHandler)).Methods("POST")
	admin.HandleFunc("/edit-books", RequireAuth(handlers.AllBooksHandler)).Methods("GET")
	admin.HandleFunc("/edit-book/{id}", RequireAuth(handlers.EditBookFormHandler)).Methods("GET")
	admin.HandleFunc("/delete-book/{id}", RequireAuth(handlers.DeleteBookFormHandler)).Methods("GET")

	return r
}
