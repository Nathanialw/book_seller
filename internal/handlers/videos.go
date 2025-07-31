package handlers

import (
	"html/template"
	"net/http"
)

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
