package db

import (
	"bookmaker.ca/internal/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
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
func InsertBook(title, author string, price float64, description, imagePath string) error {
	sql := `
		INSERT INTO books (title, author, price, description, image_path)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Exec(ctx, sql, title, author, price, description, imagePath)
	if err != nil {
		log.Printf("Failed to insert book: %v\n", err)
		return err
	}
	log.Printf("Inserted book: %s by %s\n", title, author)
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

func GetBookByID(id int) (*models.Book, error) {
	var b models.Book
	err := db.QueryRow(context.Background(), `
		SELECT id, title, author, price, description, image_path
		FROM books
		WHERE id = $1
	`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Description, &b.Image)
	if err != nil {
	}
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

func UpdateBook(id int, title, author string, price float64, desc, img string) error {
	var query string
	var err error

	if img != "" {
		query = `
			UPDATE books SET title=$1, author=$2, price=$3, description=$4, image_path=$5
			WHERE id=$6
		`
		_, err = db.Exec(ctx, query, title, author, price, desc, img, id)
	} else {
		query = `
			UPDATE books SET title=$1, author=$2, price=$3, description=$4
			WHERE id=$5
		`
		_, err = db.Exec(ctx, query, title, author, price, desc, id)
	}
	return err
}

func GetAllBooks() ([]models.Book, error) {
	rows, err := db.Query(context.Background(), `
		SELECT id, title, author, price, description, image_path
		FROM books
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		b.Image = ""
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Description, &b.Image); err != nil {
			continue
		}
		books = append(books, b)
	}

	return books, rows.Err()
}

func DeleteBook(id int) error {
	// Before deleting the book:
	// book, err := db.GetBookByID(bookID)
	// if err == nil && book.ImagePath != "" {
	// 	os.Remove("static/img/" + book.ImagePath)
	// }
	_, err := db.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	return err
}
