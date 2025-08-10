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
	Cents        int64
	Price        float64
	ProductTitle string
	VariantColor string
}
