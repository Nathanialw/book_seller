package models

type CartItem struct {
	ID        int
	VariantID int
	Quantity  int
}

type CartItems struct {
	ID       int
	Variant  Variant
	Name     string
	Quantity int
	Total    float64
}

type Cart struct {
	ID       int
	Products []CartItems
	Subtotal float64
	Tax      float64
	Total    float64
}
