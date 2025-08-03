package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"bookmaker.ca/internal/db"
)

func ProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract productID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	productID, err := strconv.Atoi(parts[len(parts)-1])

	// Handle invalid productID if needed
	if err != nil {
		http.Error(w, "Invalid Product ID", http.StatusBadRequest)
		return
	}

	// Fetch the products details from the database
	product, err := db.GetProductByID(productID)
	if err != nil {
		http.Error(w, "Failed to retrieve Product details", http.StatusInternalServerError)
		return
	}

	// Parse the template
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates//product/content.html",
		"templates//product/variant-custom.html",
		"templates//product/content-custom.html",
	)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	// Execute the template and send the response
	if err := tmpl.Execute(w, product); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func ProductListHandler(w http.ResponseWriter, r *http.Request) {
	products, err := db.GetAllProducts()
	if err != nil {
		http.Error(w, "Failed to load products", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/product/product-list.html",
	))

	if err := tmpl.Execute(w, products); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
