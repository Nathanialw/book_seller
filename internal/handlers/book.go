package handlers

import (
	"html/template"
	"net/http"

	"bookmaker.ca/internal/data"
)

//pass in the book_id of the book
func BookDetailHandler(w http.ResponseWriter, r *http.Request) {
	book_id := 0
	tmpl := template.Must(template.ParseFiles("templates/book.html"))
	tmpl.Execute(w, data.Books[book_id])
}
