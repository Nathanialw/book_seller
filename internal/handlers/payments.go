package handlers

import (
	"html/template"
	"net/http"
	"os"

	"bookmaker.ca/internal/cart"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	returnURL := r.FormValue("id")

	//TODO: set the Key as an env variable on the server
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	//TODO: //Get values from the cart
	name := "The Go Programming Language Book"
	currency := "usd"
	var quantity int64 = 1
	var amount int64 = 3499
	//TODO: links for images within the site
	image := "https://nathanial.ca/assets/images/default.png"

	if returnURL == "" {
		returnURL = "404" // default fallback
	}
	returnURL = "http://127.0.0.1:6600/product/" + returnURL

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:   stripe.String(name),
						Images: stripe.StringSlice([]string{image}),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(quantity),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		//TODO: Use real pages with full URLs
		SuccessURL: stripe.String("http://127.0.0.1:6600/success"),
		CancelURL:  stripe.String(returnURL),
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func CreateCartCheckoutSession(w http.ResponseWriter, r *http.Request) {
	cartItems := cart.CheckoutHandler(w, r)

	//TODO: set the Key as an env variable on the server
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, item := range cartItems.Products {
		amount := int64(item.Variant.Price * 100) // Stripe expects amount in cents
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("CAD"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:   stripe.String(item.Variant.Color),
					Images: stripe.StringSlice([]string{item.Variant.ImagePath}),
				},
				UnitAmount: stripe.Int64(amount),
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String("http://127.0.0.1:6600/success"),
		CancelURL:          stripe.String("http://127.0.0.1:6600/cart"),
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO:
	// insert order into the orders db
	cart.ClearCart()

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func SuccessHandler(w http.ResponseWriter, r *http.Request) {
	// books := []models.Book{
	// 	{ID: 1, Title: "Go in Action", Author: "William Kennedy", Price: 29.99, Image: "/static/img/go.jpg"},
	// 	{ID: 2, Title: "The Go Programming Language", Author: "Alan Donovan", Price: 34.99, Image: "/static/img/go2.jpg"},
	// }

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/partials/header.html",
		"templates/partials/footer.html",
		"templates/success.html",
	))

	books := 0
	tmpl.Execute(w, books)
}
