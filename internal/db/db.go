package db

import (
	"context"
	"fmt"
	"log"

	"bookmaker.ca/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var ctx = context.Background()

func InitDB() {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://bookuser:securepassword@localhost/bookmaker")
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
}

// InsertBook inserts a new book row into the books table.
func InsertBook(title, author, description string, variants []models.Variant) error {
	// First, insert into the books table
	sql := `
		INSERT INTO books (title, author, description)
		VALUES ($1, $2, $3) RETURNING id
	`
	var bookID int
	err := db.QueryRow(ctx, sql, title, author, description).Scan(&bookID)
	if err != nil {
		log.Printf("Failed to insert book: %v\n", err)
		return err
	}

	// Then insert the variants into the book_variants table
	for _, v := range variants {
		sqlVariant := `
			INSERT INTO book_variants (book_id, color, stock, price, image_path)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err := db.Exec(ctx, sqlVariant, bookID, v.Color, v.Stock, v.Price, v.ImagePath)
		if err != nil {
			log.Printf("Failed to insert variant: %v\n", err)
			return err
		}
	}
	log.Printf("Inserted book: %s by %s with %d variants\n", title, author, len(variants))
	return nil
}

func SearchBooks(query string) ([]models.Book, error) {
	rows, err := db.Query(ctx, `
        SELECT id, title, author
        FROM books
        WHERE title % $1 OR author % $1
        ORDER BY GREATEST(similarity(title, $1), similarity(author, $1)) DESC
        LIMIT 20
    `, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author); err != nil {
			continue
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

// Function to get variants by book ID
func GetVariantsByBookID(bookID int) ([]models.Variant, error) {
	var variants []models.Variant

	// Query for variants associated with the book
	rows, err := db.Query(context.Background(), `
		SELECT id, color, stock, price, image_path
		FROM book_variants
		WHERE book_id = $1
	`, bookID)
	if err != nil {
		// Handle error if variants can't be fetched
		return nil, fmt.Errorf("error fetching variants: %v", err)
	}
	defer rows.Close()

	// Scan each variant and append to the variants slice
	for rows.Next() {
		var v models.Variant
		err := rows.Scan(&v.ID, &v.Color, &v.Stock, &v.Price, &v.ImagePath)
		if err != nil {
			// Handle scanning error for variants
			return nil, fmt.Errorf("error scanning variant: %v", err)
		}
		variants = append(variants, v)
	}

	return variants, nil
}

// GetBookByID now calls GetVariantsByBookID
func GetBookByID(id int) (*models.Book, error) {
	// Initialize book
	var b models.Book

	// Query for book details
	err := db.QueryRow(context.Background(), `
		SELECT id, title, author, description
		FROM books
		WHERE id = $1
	`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Description)
	if err != nil {
		// Handle error if book is not found
		return nil, fmt.Errorf("error fetching book: %v", err)
	}

	// Get variants using the separate function
	variants, err := GetVariantsByBookID(id)
	if err != nil {
		return nil, err
	}

	// Assign the variants to the book
	b.Variants = variants

	// Return the book with variants
	return &b, nil
}

func GetAuthors() ([]string, error) {
	rows, err := db.Query(ctx, `
        SELECT DISTINCT author
        FROM books
        ORDER BY author ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
		var author string
		if err := rows.Scan(&author); err != nil {
			log.Println("Error scanning author:", err)
			continue
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}

func UpdateBookAndVariants(id int, title, author, desc string, variants []models.Variant) error {
	// Update the book information
	queryBook := `
		UPDATE books 
		SET title=$1, author=$2, description=$3
		WHERE id=$4
	`
	_, err := db.Exec(ctx, queryBook, title, author, desc, id)
	if err != nil {
		log.Printf("Failed to update book info: %v\n", err)
		return err
	}

	// Delegate variant update to separate function
	err = UpdateBookVariants(id, variants)
	if err != nil {
		return err
	}

	log.Printf("Updated book %d with %d variants\n", id, len(variants))
	return nil
}

func UpdateBookVariants(bookID int, variants []models.Variant) error {
	for _, variant := range variants {
		queryVariant := `
			UPDATE book_variants
			SET color=$1, stock=$2, price=$3, image_path=$4
			WHERE book_id=$5 AND color=$1
		`
		_, err := db.Exec(ctx, queryVariant, variant.Color, variant.Stock, variant.Price, variant.ImagePath, bookID)
		if err != nil {
			log.Printf("Failed to update variant (color: %s): %v\n", variant.Color, err)
			return err
		}
	}
	return nil
}

func UpdateBookByID(id int, title, author, desc string) error {
	// Update the book information
	queryBook := `
		UPDATE books 
		SET title=$1, author=$2, description=$3
		WHERE id=$4
	`
	_, err := db.Exec(ctx, queryBook, title, author, desc, id)
	if err != nil {
		log.Printf("Failed to update book info: %v\n", err)
		return err
	}

	return nil
}

func UpdateBookVariantByID(variant models.Variant) error {
	query := `
		UPDATE book_variants
		SET color = $1, stock = $2, price = $3, image_path = $4
		WHERE id = $5
	`
	_, err := db.Exec(ctx, query, variant.Color, variant.Stock, variant.Price, variant.ImagePath, variant.ID)
	if err != nil {
		log.Printf("Failed to update variant (id: %d): %v\n", variant.ID, err)
		return err
	}
	return nil
}

func GetAllBooks() ([]models.Book, error) {
	// Query to join books with book_variants
	rows, err := db.Query(context.Background(), `
		SELECT b.id, b.title, b.author, b.description, 
		       v.color, v.stock, v.price, v.image_path
		FROM books b
		LEFT JOIN book_variants v ON b.id = v.book_id
		ORDER BY b.id, v.color
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	var currentBook *models.Book

	// Iterate through query results
	for rows.Next() {
		var b models.Book
		var v models.Variant

		// Use pointers for nullable fields
		var color *string
		var stock *int
		var price *float64
		var imagePath *string

		// Scan the values into the appropriate variables, including pointers
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description,
			&color, &stock, &price, &imagePath)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}

		// Check if we need to create a new book or if we're still processing the same book
		if currentBook == nil || currentBook.ID != b.ID {
			// Add the previous book to the list, if any
			if currentBook != nil {
				books = append(books, *currentBook)
			}

			// Create a new book and set it as current
			currentBook = &b
			currentBook.Variants = []models.Variant{} // Initialize the variants array
		}

		// Only append the variant if it's not NULL (i.e., a valid variant)
		if color != nil {
			v.Color = *color
		}
		if stock != nil {
			v.Stock = *stock
		}
		if price != nil {
			v.Price = *price
		}
		if imagePath != nil {
			v.ImagePath = *imagePath
		}

		// Append the variant to the book's variants array if it's valid
		if v.Color != "" {
			currentBook.Variants = append(currentBook.Variants, v)
		}
	}

	// Add the last book to the list
	if currentBook != nil {
		books = append(books, *currentBook)
	}

	return books, rows.Err()
}

func DeleteVariantEntries(id int) {
	_, err := db.Exec(ctx, `DELETE FROM book_variants WHERE book_id = $1`, id)
	if err != nil {
		log.Printf("error deleting variants: %v\n", err)
	}
}

func DeleteBookEntry(id int) {
	_, err := db.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		log.Printf("error deleting book: %v\n", err)
	}
}

func DeleteVariantEntry(id int) {
	_, err := db.Exec(ctx, `DELETE FROM book_variants WHERE id = $1`, id)
	if err != nil {
		log.Printf("error deleting variants: %v\n", err)
	}
}

func DeleteBook(id int) {
	// Before deleting the book:
	// book, err := db.GetBookByID(bookID)
	// if err == nil && book.ImagePath != "" {
	// 	os.Remove("static/img/" + book.ImagePath)
	// }
	DeleteBookEntry(id)
	DeleteVariantEntries(id)
}

func InsertVariant(bookID int, color string, stock int, price float64, imagePath string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO book_variants (book_id, color, stock, price, image_path)
		VALUES ($1, $2, $3, $4, $5)
	`, bookID, color, stock, price, imagePath)
	if err != nil {
		log.Printf("InsertVariant error: %v\n", err)
	}
	return err
}

func InsertBookReturningID(title, author, description string) (int, error) {
	var id int
	sql := `
		INSERT INTO books (title, author, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := db.QueryRow(ctx, sql, title, author, description).Scan(&id)
	if err != nil {
		log.Printf("InsertBook error: %v\n", err)
		return 0, err
	}
	return id, nil
}
