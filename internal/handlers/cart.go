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

type CartItems struct {
	Variant  models.Variant
	Quantity int
	Total    float64
}

func getCartItems(r *http.Request) ([]CartItems, float64) {
	session, _ := db.Store.Get(r, "session")
	cartAny := session.Values["cart"]
	cart, _ := cartAny.([]models.CartItem)

	var total float64
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
	return products, total
}

func calcTax(total float64) (float64, float64) {
	const GST = 0.05
	tax := math.Round(total*GST*100) / 100
	subtotal := total + tax

	return subtotal, tax
}

func CartHandler(w http.ResponseWriter, r *http.Request) {

	products, total := getCartItems(r)
	subtotal, tax := calcTax(total)

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
	idStr := r.FormValue("increment")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, _ := db.Store.Get(r, "session")
	cartAny := session.Values["cart"]
	cart, ok := cartAny.([]models.CartItem)
	if !ok {
		http.Error(w, "Cart not found", http.StatusInternalServerError)
		return
	}

	// Correctly increment quantity by index
	for i := range cart {
		if cart[i].VariantID == id {
			// TODO: Check against in stock in the db
			cart[i].Quantity++
			break
		}
	}

	session.Values["cart"] = cart
	session.Save(r, w)

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func DecrementItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("decrement")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, _ := db.Store.Get(r, "session")
	cartAny := session.Values["cart"]
	cart, ok := cartAny.([]models.CartItem)
	if !ok {
		http.Error(w, "Cart not found", http.StatusInternalServerError)
		return
	}

	// Decrement quantity or remove item entirely
	for i := range cart {
		if cart[i].VariantID == id {
			if cart[i].Quantity > 1 {
				cart[i].Quantity--
			} else {
				// Remove item from slice
				cart = append(cart[:i], cart[i+1:]...)
			}
			break
		}
	}

	session.Values["cart"] = cart
	session.Save(r, w)

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func RemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("remove-item")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	session, _ := db.Store.Get(r, "session")
	cartAny := session.Values["cart"]
	cart, ok := cartAny.([]models.CartItem)
	if !ok {
		http.Error(w, "Cart not found", http.StatusInternalServerError)
		return
	}

	// Decrement quantity or remove item entirely
	for i := range cart {
		if cart[i].VariantID == id {
			cart = append(cart[:i], cart[i+1:]...)
			break
		}
	}

	session.Values["cart"] = cart
	session.Save(r, w)

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}
