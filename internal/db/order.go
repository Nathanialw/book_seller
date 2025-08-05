package db

import (
	"log"

	"bookmaker.ca/internal/models"
)

func SearchOrders(email string, orderNumber string) (models.Order, error) {
	var order models.Order

	row := db.QueryRow(ctx, `
        SELECT id, email, number
        FROM orders
        WHERE email % $1 AND number::text % $2
        LIMIT 1
    `, email, orderNumber)

	err := row.Scan(&order.ID, &order.Email, &order.Number)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func InsertOrder(email string, orderNumber int, products []models.Product) int {
	var id int
	sql := `
		INSERT INTO orders (email, orderNumber)
		VALUES ($1, $2)
		RETURNING id
	`
	err := db.QueryRow(ctx, sql, email, orderNumber).Scan(&id)
	if err != nil {
		log.Printf("InsertProduct error: %v\n", err)
		return 0
	}
	return id
}

func DeletOrder(email string, orderNumber int, products []models.Product) {

}
