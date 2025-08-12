package models

type Product struct {
	ID          int
	Title       string
	Author      string
	Description string
	LowestPrice float64
	Type0       string

	Variants []Variant
}

type Variant struct {
	ID         int
	Product_ID int //`foreign:Product(ID)` //or just Product_ID
	Color      string
	Stock      int
	Cents      int64
	Price      float64
	ImagePath  string
}

// //size
// Pages  string
// Width  string
// Height string
// Length string

// //materials
// Cover string
// Paper string

// //process
// Binding string
