package db

import (
	"log"

	"github.com/nathanialw/ecommerce/pkg/models"
)

func SearchOrders(email string, orderNumber string) (models.Order, error) {
	// TODO:
	// Check order number for ORD- prepend

	var order models.Order

	row := db.QueryRow(ctx, `
        SELECT id, order_id, email, address, city, postal_code, country, created_at
        FROM orders
    	WHERE email = $1 AND order_id = $2
        LIMIT 1
    `, email, orderNumber)

	err := row.Scan(
		&order.ID,
		&order.OrderNumber,
		&order.Email,
		&order.Address,
		&order.City,
		&order.PostalCode,
		&order.Country,
		&order.CreatedAt,
	)
	if err != nil {
		return models.Order{}, err
	}
	// Get the order items
	rows, err := db.Query(ctx, `
        SELECT product_variant_id, quantity, price, product_title, variant_color 
        FROM order_items
    	WHERE order_id = $1
    `, order.ID)
	if err != nil {
		return models.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.VariantID,
			&item.Quantity,
			&item.Cents,
			&item.ProductTitle,
			&item.VariantColor,
		)
		if err != nil {
			return models.Order{}, err
		}
		item.Price = float64(item.Cents) / 100.0
		order.Products = append(order.Products, item)
	}

	return order, nil
}

func InsertOrder(orderNumber, email, address, city, postalCode, country string, items []models.OrderItem) (orderID int, err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Printf("1Failed to create order: %v", err)
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (order_id, email, address, city, postal_code, country) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		orderNumber, email, address, city, postalCode, country,
	).Scan(&orderID)
	if err != nil {
		log.Printf("2Failed to create order: %v", err)
		return 0, err
	}

	for _, item := range items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_variant_id, quantity, price, product_title, variant_color) VALUES ($1, $2, $3, $4, $5, $6)`,
			orderID, item.VariantID, item.Quantity, item.Cents, item.ProductTitle, item.VariantColor,
		)
		if err != nil {
			log.Printf("3Failed to create order: %v", err)
			return 0, err
		}
	}

	return orderID, nil
}

func DeletOrder(email string, orderNumber int, products []models.Product) {

}
