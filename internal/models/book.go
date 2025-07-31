package models

type Book struct {
	ID          int
	Title       string
	Author      string
	Description string
	Price       float64
	Image       string

	//size
	Pages  string
	Width  string
	Height string
	Length string

	//materials
	Cover string
	Paper string

	//process
	Binding string
}
