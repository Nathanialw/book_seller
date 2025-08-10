package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/Nathanialw/ecommerce/internal/db"
	"github.com/Nathanialw/ecommerce/internal/models"
)

func GenerateShortOrderID() string {
	bytes := make([]byte, 5) // 10 hex characters = 40 bits of entropy
	rand.Read(bytes)
	id := "ORD-" + hex.EncodeToString(bytes)
	println("generating order number: ", id)
	return id
}

// TODO:
func CreateOrder(orderNumber, email, address, city, postalCode, country string, items []models.OrderItem) {
	db.InsertOrder(orderNumber, email, address, city, postalCode, country, items)
	println("NOT IMPLEMENTED creating order: ", orderNumber)
}

// TODO:
func SaveShippingAddress(name, line, city, postalCode, country string) {
	fmt.Println("NOT IMPLEMENTED saving shipping address: ", name, line, city, postalCode, country)
}

// TODO:
func EmailOrderDetails(email string) {
	fmt.Println("NOT IMPLEMENTED emailing order details to: ", email)
}
