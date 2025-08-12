package models

import "time"

type CartItem struct {
	ID         int
	Variant_ID int
	Name       string
	Quantity   int
	Total      float64
	CreatedAt  time.Time

	//not to be  stored in db
	Variant Variant
}

type Cart struct {
	ID        int
	Subtotal  float64
	Tax       float64
	Total     float64
	CreatedAt time.Time

	//not to be  stored in db
	Products []CartItem
}
