package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"bookmaker.ca/internal/db"
)

func SearchProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the 'q' query parameter from URL
	query := r.URL.Query().Get("q")

	if query == "" {
		http.Error(w, "Query parameter 'q' is missing", http.StatusBadRequest)
		return
	}

	// Call your db.SearchDB(query) but you need to adjust it to return results instead of just printing

	results, err := db.SearchProducts(query) // See next note about this function
	if err != nil {
		http.Error(w, "Search error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render results (e.g., with template)

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/search-results.html",
	))

	tmpl.Execute(w, results)
}

func SearchOrdersHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	orderNumber := r.URL.Query().Get("order-number")

	println(email)
	println(orderNumber)

	if email == "" || orderNumber == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Email and Order Number are required",
		})
		return
	}

	results, err := db.SearchOrders(email, orderNumber)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "No orders found for that email and order number.",
		})
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/order/order-results.html",
	))

	err = tmpl.Execute(w, results) // or wrap in struct if template needs that
	if err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
