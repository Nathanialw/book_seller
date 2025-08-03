package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"bookmaker.ca/internal/db"
)

func BookDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract bookID from the URL path
	parts := strings.Split(r.URL.Path, "/")
	bookID, err := strconv.Atoi(parts[len(parts)-1])

	// Handle invalid bookID if needed
	if err != nil {
		http.Error(w, "Invalid Book ID", http.StatusBadRequest)
		return
	}

	// Fetch the book details from the database
	book, err := db.GetBookByID(bookID)
	if err != nil {
		http.Error(w, "Failed to retrieve book details", http.StatusInternalServerError)
		return
	}

	// Parse the template
	tmpl, err := template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/product.html",
		"templates/variant-custom.html",
		"templates/book.html",
	)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	// Execute the template and send the response
	if err := tmpl.Execute(w, book); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func BookListHandler(w http.ResponseWriter, r *http.Request) {
	books, err := db.GetAllBooks()
	if err != nil {
		http.Error(w, "Failed to load books", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/booklist.html",
	))

	if err := tmpl.Execute(w, books); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
