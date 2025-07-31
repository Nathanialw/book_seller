package handlers

import (
	"html/template"
	"net/http"

	"bookmaker.ca/internal/cache"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/home.html",
	))

	authors := cache.GetAuthors() // This should return []string or []Author

	data := struct {
		Authors []string
	}{
		Authors: authors,
	}

	tmpl.Execute(w, data)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/about.html",
	))
	d := 0
	tmpl.Execute(w, d)
}

func VideosHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/videos.html",
	))
	d := 0
	tmpl.Execute(w, d)
}
