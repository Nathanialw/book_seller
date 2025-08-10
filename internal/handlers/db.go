package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"bookmaker.ca/internal/db"
)

func SearchOrdersHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	orderNumber := r.URL.Query().Get("order-number")

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
