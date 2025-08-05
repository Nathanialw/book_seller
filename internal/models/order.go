package models

type Order struct {
	ID       int
	Number   int
	Email    string
	Products []Product
}
