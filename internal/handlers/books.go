package handlers

import (
	"html/template"
	"net/http"

	"bookmaker.ca/internal/data"
)

func BookListHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/booklist.html",
	))

	tmpl.Execute(w, data.Books)
}
