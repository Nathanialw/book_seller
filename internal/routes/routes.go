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

	r.HandleFunc("/products", handlers.ProductListHandler).Methods("GET")
	r.HandleFunc("/product/{id}", handlers.ProductDetailHandler).Methods("GET")
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
	admin.HandleFunc("/add-product", RequireAuth(handlers.AddProductForm)).Methods("GET")
	admin.HandleFunc("/add-product", RequireAuth(handlers.AddProductHandler)).Methods("POST")
	admin.HandleFunc("/update-product", RequireAuth(handlers.UpdateProductHandler)).Methods("POST")
	admin.HandleFunc("/edit-products", RequireAuth(handlers.EditAllProductssHandler)).Methods("GET")
	admin.HandleFunc("/edit-product/{id}", RequireAuth(handlers.EditProductFormHandler)).Methods("GET")
	admin.HandleFunc("/delete-product/{id}", RequireAuth(handlers.DeleteProductFormHandler)).Methods("GET")

	return r
}
