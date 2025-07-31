package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"bookmaker.ca/internal/db"
)

//TODO: add validation for admin login
func AdminLoginGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/login.html",
	))

	d := 0
	tmpl.Execute(w, d)
}

func AddBookForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/add_book.html",
	))

	d := 0
	tmpl.Execute(w, d)
}

func AddBookSubmit(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	author := r.FormValue("author")
	priceStr := r.FormValue("price")
	description := r.FormValue("description")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	err = db.InsertBook(title, author, price, description)
	if err != nil {
		http.Error(w, "Failed to insert", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/add-book", http.StatusSeeOther)
}
