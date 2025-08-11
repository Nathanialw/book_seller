package db

import (
	"context"
	"log"

	"github.com/nathanialw/ecommerce/pkg/models"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var ctx = context.Background()
var Store = sessions.NewCookieStore([]byte("super-secret-key")) // use a strong key in production

func InitDB() *pgxpool.Pool {
	var err error
	db, err = pgxpool.New(context.Background(), "postgres://admin:securepassword@localhost/ecommerce")
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	return db
}

func GetCache() ([]string, error) {
	rows, err := db.Query(ctx, `
        SELECT DISTINCT author
        FROM products
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

func UpdateProductAndVariants(id int, title, author, desc string, variants []models.Variant) error {
	// Update the product information
	queryProductID := `
		UPDATE products 
		SET title=$1, author=$2, description=$3
		WHERE id=$4
	`
	_, err := db.Exec(ctx, queryProductID, title, author, desc, id)
	if err != nil {
		log.Printf("Failed to update product info: %v\n", err)
		return err
	}

	// Delegate variant update to separate function
	err = UpdateProductVariants(id, variants)
	if err != nil {
		return err
	}

	log.Printf("Updated product %d with %d variants\n", id, len(variants))
	return nil
}
