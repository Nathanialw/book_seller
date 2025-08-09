package models

import "time"

type Order struct {
	ID          int
	OrderNumber string
	Email       string
	Address     string
	City        string
	PostalCode  string
	Country     string
	CreatedAt   time.Time
	Products    []OrderItem
}

type OrderItem struct {
	VariantID    int
	Quantity     int
	Price        int64
	ProductTitle string
	VariantColor string
}
