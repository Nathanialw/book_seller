package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateShortOrderID() string {
	bytes := make([]byte, 5) // 10 hex characters = 40 bits of entropy
	rand.Read(bytes)
	id := "ORD-" + hex.EncodeToString(bytes)
	println("generating order number: ", id)
	return id
}

// TODO:
func CreateOrder(orderID string) {
	println("creating order: ", orderID)
}

// TODO:
func SaveShippingAddress(name, line, city, postalCode, country string) {
	fmt.Println("saving shipping address: ", name, line, city, postalCode, country)
}

// TODO:
func EmailOrderDetails(email string) {
	fmt.Println("emailing order details to: ", email)
}
