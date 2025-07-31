package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"bookmaker.ca/internal/db"
)

//pass in the book_id of the book
func BookDetailHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	bookID, err := strconv.Atoi(parts[len(parts)-1])

	if err != nil || bookID <= 0 {
		http.NotFound(w, r)
		return
	}

	book, err := db.GetBookByID(bookID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/book.html",
	))

	if err := tmpl.Execute(w, book); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
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
