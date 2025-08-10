package services

import (
	"math"
	"net/http"

	"github.com/nathanialw/ecommerce/internal/db"
	"github.com/nathanialw/ecommerce/pkg/models"
)

func GetCartItems(r *http.Request) ([]models.CartItems, float64) {
	session, _ := db.Store.Get(r, "session")
	cartAny := session.Values["cart"]
	cart, _ := cartAny.([]models.CartItem)

	var total float64
	var products []models.CartItems
	for _, item := range cart {
		variant, err := db.GetVariantByID(item.VariantID)
		product, _ := db.GetProductByID(variant.ID)
		if err == nil {
			products = append(products, models.CartItems{
				Variant:  variant,
				Quantity: item.Quantity,
				Total:    variant.Price * float64(item.Quantity),
				Name:     product.Title,
			})
			total += variant.Price * float64(item.Quantity)
		}
	}
	return products, total
}

func CalcTax(total float64) (float64, float64) {
	const GST = 0.05
	tax := math.Round(total*GST*100) / 100
	subtotal := total + tax

	return subtotal, tax
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) models.Cart {
	products, total := GetCartItems(r)
	subtotal, tax := CalcTax(total)

	data := models.Cart{
		Products: products,
		Subtotal: subtotal,
		Tax:      tax,
		Total:    total,
	}

	return data
}

// TODO:
func ClearCart() {
	println("NOT IMPLEMENTED clearing cart")
}
