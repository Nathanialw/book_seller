package handlers

import (
	"html/template"
	"net/http"

	"bookmaker.ca/internal/cache"
)

func loggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		return false
	}
	// TODO:
	// validate session value
	return true
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/home.html",
	))

	authors := cache.GetCache() // This should return []string or []Author
	loggedIn := loggedIn(r)

	data := struct {
		Authors  []string
		LoggedIn bool
	}{
		Authors:  authors,
		LoggedIn: loggedIn,
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

func BlogsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/blog.html",
	))
	d := 0
	tmpl.Execute(w, d)
}
