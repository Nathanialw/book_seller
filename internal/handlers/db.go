package handlers

import (
	"html/template"
	"net/http"

	"bookmaker.ca/internal/db"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
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
