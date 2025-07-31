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
	book_id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || book_id < 0 || book_id >= len(db.Books) {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/book.html",
	))

	tmpl.Execute(w, db.Books[book_id])
}

func BookListHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/booklist.html",
	))

	tmpl.Execute(w, db.Books)
}
