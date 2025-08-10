package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/nathanialw/ecommerce/internal/db"
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
		"templates/partials/search.html",
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

	for i := range products {
		if len(products[i].Variants) == 0 {
			continue // no variants, skip
		}
		price := products[i].Variants[0].Cents
		for _, variant := range products[i].Variants[1:] {
			if variant.Cents < price {
				price = variant.Cents
			}
		}
		products[i].LowestPrice = float64(price) / 100
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/search.html",
		"templates/partials/footer.html",
		"templates/product/product-grid.html",
		"templates/product/product-list.html",
	))

	if err := tmpl.Execute(w, products); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func SearchProductsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the 'q' query parameter from URL
	query := r.URL.Query().Get("q")

	if query == "" {
		http.Error(w, "Query parameter 'q' is missing", http.StatusBadRequest)
		return
	}

	// Call your db.SearchDB(query) but you need to adjust it to return results instead of just printing

	products, err := db.SearchProducts(query) // See next note about this function
	if err != nil {
		http.Error(w, "Search error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range products {
		if len(products[i].Variants) == 0 {
			continue // no variants, skip
		}
		price := products[i].Variants[0].Cents
		for _, variant := range products[i].Variants[1:] {
			if variant.Cents < price {
				price = variant.Cents
			}
		}
		products[i].LowestPrice = float64(price) / 100
	}

	// Render results (e.g., with template)
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/search.html",
		"templates/partials/footer.html",
		"templates/product/product-grid.html",
		"templates/product/search-results.html",
	))

	tmpl.Execute(w, products)
}
