package models

type Book struct {
	ID          int
	Title       string
	Author      string
	Description string

	Variants []Variant

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

}

type Variant struct {
	ID        int
	Color     string
	Stock     int
	Price     float64
	ImagePath string
}
