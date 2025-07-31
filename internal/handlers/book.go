package handlers

import (
	"html/template"
	"net/http"

	"bookmaker.ca/internal/models"
)

func BookListHandler(w http.ResponseWriter, r *http.Request) {
	books := []models.Book{
		{ID: 1, Title: "Go in Action", Author: "William Kennedy", Price: 29.99, Image: "/static/img/go.jpg"},
		{ID: 2, Title: "The Go Programming Language", Author: "Alan Donovan", Price: 34.99, Image: "/static/img/go2.jpg"},
	}

	tmpl := template.Must(template.ParseFiles("templates/booklist.html"))
	tmpl.Execute(w, books)
}
