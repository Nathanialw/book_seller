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
	//not to be  stored in db
	Products []OrderItem
}

type OrderItem struct {
	ID           int
	Order_ID     int
	Variant_ID   int
	Quantity     int
	Cents        int64
	ProductTitle string
	VariantColor string
	CreatedAt    time.Time
	//not to be  stored in db
	Price float64
}
