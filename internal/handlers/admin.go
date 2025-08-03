package handlers

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"bookmaker.ca/internal/admin"
	"bookmaker.ca/internal/cache"
	"bookmaker.ca/internal/db"
	"bookmaker.ca/internal/models"
)

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/admin/login.html",
	))

	d := 0
	tmpl.Execute(w, d)
}

func validLogin(username, password string) bool {
	// For demo purposes only — replace with real DB check
	const adminUser = "admin"
	const adminPass = "securepassword123"

	return username == adminUser && password == adminPass
}

func AdminLoginValidateHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	if validLogin(r.FormValue("username"), r.FormValue("password")) {
		// Set cookie (you could use secure sessions like gorilla/sessions)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "some-auth-token-or-username",
			Path:     "/",
			HttpOnly: true,
			// TODO:
			// Secure: true, // use in production
		})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid login", http.StatusUnauthorized)
}

func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Expire the cookie
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/admin/admin.html",
	))

	d := struct {
		LoggedIn bool
	}{
		LoggedIn: true,
	}

	tmpl.Execute(w, d)
}

func AdminBlogHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/admin/blogs.html",
	))

	d := struct {
		LoggedIn bool
	}{
		LoggedIn: true,
	}
	tmpl.Execute(w, d)
}

func AdminVideosHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/admin/videos.html",
	))

	d := struct {
		LoggedIn bool
	}{
		LoggedIn: true,
	}
	tmpl.Execute(w, d)
}

func AddBookForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/admin/add_book.html",
	))

	d := struct {
		LoggedIn bool
	}{
		LoggedIn: true,
	}
	tmpl.Execute(w, d)
}

func AddBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get form values for the book
	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	// Insert the book into the books table (no variants yet)
	bookID, err := db.InsertBookReturningID(title, author, description)
	if err != nil {
		http.Error(w, "Failed to insert book", http.StatusInternalServerError)
		return
	}

	// Now process the variants
	colors := r.Form["color"] // Array of colors
	stocks := r.Form["stock"] // Array of stock values
	prices := r.Form["price"]
	imageFiles := r.MultipartForm.File["variant_image"] // Array of variant images

	// We need to ensure all arrays have the same length
	if len(colors) != len(stocks) || len(colors) != len(imageFiles) {
		println("Mismatch in variant data")
	} else {
		// Insert each variant using the InsertVariant function
		for i := range colors {
			color := colors[i]
			stock, _ := strconv.Atoi(stocks[i])
			imagePath := ""

			price, err := strconv.ParseFloat(prices[i], 64)
			if err != nil {
				http.Error(w, "Invalid price", http.StatusBadRequest)
				return
			}

			// Override only if a new file was uploaded
			if i < len(imageFiles) {
				file, err := imageFiles[i].Open()
				if err != nil {
					log.Printf("Failed to open file %d: %v", i, err)
				} else {
					defer file.Close()
					imagePath = imageFiles[i].Filename
					savePath := "static/img/" + imagePath

					dst, err := os.Create(savePath)
					if err != nil {
						log.Printf("Failed to create file %d: %v", i, err)
					} else {
						defer dst.Close()
						_, err = io.Copy(dst, file)
						if err != nil {
							log.Printf("Failed to write file %d: %v", i, err)
						}
					}
				}
			}

			// Insert the variant into the book_variants table
			err = db.InsertVariant(bookID, color, stock, price, imagePath)
			if err != nil {
				http.Error(w, "Failed to insert variant", http.StatusInternalServerError)
				return
			}
		}

	}

	cache.UpdateAuthors()

	// Redirect back to the admin page after successful book and variant creation
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
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
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/admin/edit_book.html",
	))

	d := struct {
		LoggedIn bool
		Book     models.Book
	}{
		LoggedIn: true,
		Book:     *book,
	}

	tmpl.Execute(w, d)
}

func UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	idStr := r.FormValue("id")

	switch {
	case action == "update":
		println("update...")
		admin.UpdateBook(w, r)
		http.Redirect(w, r, "/admin/edit-books", http.StatusSeeOther)
	case strings.HasPrefix(action, "remove_variant-"):
		println("remove variant...")
		variantIDStr := strings.TrimPrefix(action, "remove_variant-")
		DeleteVariantFormHandler(w, r, variantIDStr)
		http.Redirect(w, r, "/admin/edit-book/"+idStr, http.StatusSeeOther)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}

func EditAllBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := db.GetAllBooks()

	if err != nil {
		http.Error(w, "Failed to fetch books", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/admin/header.html",
		"templates/partials/footer.html",
		"templates/admin/edit_books.html",
	))

	d := struct {
		LoggedIn bool
		Books    []models.Book
	}{
		LoggedIn: true,
		Books:    books,
	}
	tmpl.Execute(w, d)
}

func DeleteBookFormHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/delete-book/")
	bookID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	db.DeleteBook(bookID)

	log.Printf("Deleted book with ID %d", bookID)
	cache.UpdateAuthors()
	http.Redirect(w, r, "/admin/edit-books", http.StatusSeeOther)
}

func DeleteVariantFormHandler(w http.ResponseWriter, r *http.Request, variantIDStr string) {
	variantID, err := strconv.Atoi(variantIDStr)
	if err != nil {
		http.Error(w, "Invalid variant ID", http.StatusBadRequest)
		return
	}

	println("deleting variant")
	db.DeleteVariantEntry(variantID)
}
