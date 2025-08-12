package models

import "time"

type CartItem struct {
	ID        int
	VariantID int
	Quantity  int
	CreatedAt time.Time
}

type CartItems struct {
	ID        int
	Variant   Variant
	Name      string
	Quantity  int
	Total     float64
	CreatedAt time.Time
}

type Cart struct {
	ID        int
	Products  []CartItems
	Subtotal  float64
	Tax       float64
	Total     float64
	CreatedAt time.Time
}
