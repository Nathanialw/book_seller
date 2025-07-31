package db

import (
	"bookmaker.ca/internal/models"
	"context"
	"fmt"
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
func InsertBook(title, author string, price float64, description string) error {
	// Prepare SQL insert statement with placeholders
	sql := `
		INSERT INTO books (title, author, price, description)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(ctx, sql, title, author, price, description)
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
		SELECT id, title, author, price, description
		FROM books
		WHERE id = $1
	`, id).Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Description)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}
	return &b, nil
}

func GetAllBooks() ([]models.Book, error) {
	rows, err := db.Query(context.Background(), `
		SELECT id, title, author, price, description
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
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Description); err != nil {
			continue
		}
		books = append(books, b)
	}
	return books, rows.Err()
}
