package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bookmaker.ca/internal/cache"
	"bookmaker.ca/internal/db"
)

//TODO: add validation for admin login
func AdminLoginGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/admin/login.html",
	))

	d := 0
	tmpl.Execute(w, d)
}

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/admin/admin.html",
	))

	d := 0
	tmpl.Execute(w, d)
}

func AddBookForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/admin/add_book.html",
	))

	d := 0
	tmpl.Execute(w, d)
}

func AddBookHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // max 10MB file
	if err != nil {
		http.Error(w, "Cannot parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")
	priceStr := r.FormValue("price")
	description := r.FormValue("description")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	// Handle file upload
	file, header, err := r.FormFile("cover")
	var imagePath string
	if err == nil {
		defer file.Close()
		filename := filepath.Base(header.Filename)
		imagePath = filename
		out, err := os.Create("static/img/" + imagePath)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
	}

	// Insert book into database (with image path)
	err = db.InsertBook(title, author, price, description, imagePath)
	if err != nil {
		http.Error(w, "Failed to insert book", http.StatusInternalServerError)
		return
	}

	cache.UpdateAuthors()
	http.Redirect(w, r, "/admin/add-book", http.StatusSeeOther)
}

func EditBookFormHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/edit-book/")
	bookID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	book, err := db.GetBookByID(bookID)

	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/admin/edit_book.html",
	))
	tmpl.Execute(w, book)
}

func UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	title := r.FormValue("title")
	author := r.FormValue("author")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	description := r.FormValue("description")

	// Handle image
	file, handler, err := r.FormFile("image")
	var imagePath string
	if err == nil {
		defer file.Close()
		imagePath = fmt.Sprintf("%s", handler.Filename)
		dst, err := os.Create("static/img/" + imagePath)
		if err == nil {
			defer dst.Close()
			io.Copy(dst, file)
		}
	}

	err = db.UpdateBook(id, title, author, price, description, imagePath)
	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	cache.UpdateAuthors()
	http.Redirect(w, r, "/admin/edit-books", http.StatusSeeOther)
}

func AllBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := db.GetAllBooks()

	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/admin/edit_books.html",
	))
	tmpl.Execute(w, books)
}

//TODO: Add delete logic
func DeleteBookFormHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/delete-book/")
	bookID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	err = db.DeleteBook(bookID)
	if err != nil {
		http.Error(w, "Failed to delete book", http.StatusInternalServerError)
		return
	}

	log.Printf("Deleted book with ID %d", bookID)
	cache.UpdateAuthors()
	http.Redirect(w, r, "/admin/edit-books", http.StatusSeeOther)
}
