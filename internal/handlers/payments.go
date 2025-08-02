package handlers

import (
	"html/template"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	//TODO: set the Key as an env variable on the server
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	returnURL := r.FormValue("id")
	//TODO: //use the ID to get the value of the book
	var amount int64 = 3499

	if returnURL == "" {
		returnURL = "404" // default fallback
	}
	returnURL = "http://127.0.0.1:6600/book/" + returnURL

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("The Go Programming Language Book"),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
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

func SuccessHandler(w http.ResponseWriter, r *http.Request) {
	// books := []models.Book{
	// 	{ID: 1, Title: "Go in Action", Author: "William Kennedy", Price: 29.99, Image: "/static/img/go.jpg"},
	// 	{ID: 2, Title: "The Go Programming Language", Author: "Alan Donovan", Price: 34.99, Image: "/static/img/go2.jpg"},
	// }

	tmpl := template.Must(template.ParseFiles("templates/success.html"))
	books := 0
	tmpl.Execute(w, books)
}
