package models

type CartItem struct {
	VariantID int
	Quantity  int
}

type CartItems struct {
	Variant  Variant
	Quantity int
	Total    float64
}

type Cart struct {
	Products []CartItems
	Subtotal float64
	Tax      float64
	Total    float64
}
