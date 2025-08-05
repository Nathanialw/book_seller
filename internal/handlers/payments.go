package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"io/ioutil"

	"bookmaker.ca/internal/core"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
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

	orderID := core.GenerateShortOrderID()

	params := &stripe.CheckoutSessionParams{
		ShippingAddressCollection: &stripe.CheckoutSessionShippingAddressCollectionParams{
			AllowedCountries: stripe.StringSlice([]string{"CA", "US"}), // or any countries you want
		},

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
		Metadata:         map[string]string{"order_id": orderID},
		SuccessURL:       stripe.String("http://127.0.0.1:6600/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:        stripe.String(returnURL),
		CustomerCreation: stripe.String("always"),
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func CreateCartCheckoutSession(w http.ResponseWriter, r *http.Request) {
	cartItems := core.CheckoutHandler(w, r)

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

	orderID := core.GenerateShortOrderID()

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Metadata:           map[string]string{"order_id": orderID},
		LineItems:          lineItems,
		ShippingAddressCollection: &stripe.CheckoutSessionShippingAddressCollectionParams{
			AllowedCountries: stripe.StringSlice([]string{"CA", "US"}),
		},
		Mode:             stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:       stripe.String("http://127.0.0.1:6600/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:        stripe.String("http://127.0.0.1:6600/cart"),
		CustomerCreation: stripe.String("always"),
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔔 Webhook received")

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		fmt.Println("❌ Signature verification failed:", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	fmt.Println("Event Type:", event.Type)

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			http.Error(w, "Failed to parse webhook", http.StatusBadRequest)
			return
		}

		if session.CollectedInformation.ShippingDetails != nil {
			address := session.CollectedInformation.ShippingDetails.Address
			name := session.CollectedInformation.ShippingDetails.Name

			fmt.Println("📦 Shipping to:", name)
			fmt.Println("📍 Address:", address.Line1, address.City, address.PostalCode, address.Country)

			// Use this info to store in your DB:
			core.SaveShippingAddress(name, address.Line1, address.City, address.PostalCode, address.Country)

		}

		email := session.CustomerDetails.Email
		fmt.Println("📧 Email:", email)

		order_id := session.Metadata["order_id"]

		// TODO: Match session.ID or customer ID to user/cart
		core.CreateOrder(order_id)
		// and then clear the cart or mark order as paid
		core.ClearCart()

		core.EmailOrderDetails(email)

		fmt.Println("✅ Payment successful for session:", session.ID)
		// e.g., cart.ClearCart(userID) or update order status in DB
	}

	w.WriteHeader(http.StatusOK)
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
