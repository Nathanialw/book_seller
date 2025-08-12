package models

import "time"

type Product struct {
	ID          int
	Title       string
	Author      string
	Description string
	CreatedAt   time.Time
	//not to be  stored in db
	LowestPrice float64
	Type0       string

	Variants []Variant
}

type Variant struct {
	ID         int
	Product_ID int //`foreign:Product(ID)` //or just Product_ID
	Color      string
	ImagePath  string
	Cents      int64
	Stock      int
	CreatedAt  time.Time
	//not to be  stored in db
	Price float64
}
