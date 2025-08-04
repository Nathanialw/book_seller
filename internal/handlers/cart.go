package handlers

import (
	"html/template"
	"math"
	"net/http"
	"strconv"

	"bookmaker.ca/internal/db"
	"bookmaker.ca/internal/models"
)

func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := db.Store.Get(r, "session")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Retrieve or create cart
	var cart []models.CartItem
	if v, ok := session.Values["cart"].([]models.CartItem); ok {
		cart = v
	}

	// Check if variant already in cart
	found := false
	for i, item := range cart {
		if item.VariantID == id {
			cart[i].Quantity++
			found = true
			break
		}
	}

	if !found {
		cart = append(cart, models.CartItem{VariantID: id, Quantity: 1})
	}

	session.Values["cart"] = cart
	session.Save(r, w)

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func CartHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := db.Store.Get(r, "session")

	cartAny := session.Values["cart"]
	cart, ok := cartAny.([]models.CartItem)
	if !ok || len(cart) == 0 {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	type CartItems struct {
		Variant  models.Variant
		Quantity int
		Total    float64
	}

	var subtotal float64
	var total float64
	var tax float64
	var products []CartItems
	for _, item := range cart {
		variant, err := db.GetVariantByID(item.VariantID)
		if err == nil {
			products = append(products, CartItems{
				Variant:  variant,
				Quantity: item.Quantity,
				Total:    variant.Price * float64(item.Quantity),
			})
			total += variant.Price * float64(item.Quantity)
		}
	}

	const GST = 0.05

	tax = math.Round(total*GST*100) / 100
	subtotal = total + tax

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/product/cart.html",
	))

	data := struct {
		Products []CartItems
		Subtotal float64
		Tax      float64
		Total    float64
	}{
		Products: products,
		Subtotal: subtotal,
		Tax:      tax,
		Total:    total,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	// add up cart
}

func IncrementItemHandler(w http.ResponseWriter, r *http.Request) {
}

func DecrementItemHandler(w http.ResponseWriter, r *http.Request) {
}

func RemoveItemHandler(w http.ResponseWriter, r *http.Request) {
}
