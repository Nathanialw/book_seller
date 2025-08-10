package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/Nathanialw/ecommerce/internal/models"
	"github.com/Nathanialw/ecommerce/internal/services"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	sessionpkg "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	returnURL := r.FormValue("id")

	//TODO: set the Key as an env variable on the server
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	//TODO: //Get values from the cart
	name := "The Go Programming Language Book"
	color := "red"
	var quantity int64 = 1
	var amount int64 = 3499
	//TODO: links for images within the site
	image := "https://nathanial.ca/assets/images/default.png"

	if returnURL == "" {
		returnURL = "404" // default fallback
	}
	returnURL = "http://127.0.0.1:6600/product/" + returnURL

	var lineItems []*stripe.CheckoutSessionLineItemParams
	lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String("CAD"),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name:        stripe.String(name),
				Images:      stripe.StringSlice([]string{image}),
				Description: stripe.String(fmt.Sprintf("Variant ID: %s", color)),
			},
			UnitAmount: stripe.Int64(amount),
		},
		Quantity: stripe.Int64(quantity),
	})

	params := params(lineItems, returnURL)

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func CreateCartCheckoutSession(w http.ResponseWriter, r *http.Request) {
	cartItems := services.CheckoutHandler(w, r)

	//TODO: set the Key as an env variable on the server
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, item := range cartItems.Products {
		amount := int64(item.Variant.Cents) // Stripe expects amount in cents
		// TODO: needs the web address of the image asset
		imgPath := "https://nathanial.ca/assets/images/default.png" // + item.Variant.ImagePath

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("CAD"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:        stripe.String(item.Name),
					Images:      stripe.StringSlice([]string{imgPath}),
					Description: stripe.String(fmt.Sprintf("Variant: %s", item.Variant.Color)),
					Metadata: map[string]string{
						"variant_id": fmt.Sprintf("%d", item.Variant.ID),
					},
				},
				UnitAmount: stripe.Int64(amount),
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	params := params(lineItems, "http://127.0.0.1:6600/cart")

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func params(lineItems []*stripe.CheckoutSessionLineItemParams, cancelURL string) *stripe.CheckoutSessionParams {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.CheckoutSessionParams{
		ShippingAddressCollection: &stripe.CheckoutSessionShippingAddressCollectionParams{
			AllowedCountries: stripe.StringSlice([]string{"CA", "US"}),
		},
		ShippingOptions: []*stripe.CheckoutSessionShippingOptionParams{
			{
				ShippingRateData: &stripe.CheckoutSessionShippingOptionShippingRateDataParams{
					DisplayName: stripe.String("Standard Shipping"),
					Type:        stripe.String("fixed_amount"),
					FixedAmount: &stripe.CheckoutSessionShippingOptionShippingRateDataFixedAmountParams{
						Amount:   stripe.Int64(1500), // in cents
						Currency: stripe.String("cad"),
					},
				},
			},
		},

		AllowPromotionCodes: stripe.Bool(true),

		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String("http://127.0.0.1:6600/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:          stripe.String(cancelURL),
		// ReturnURL:        stripe.String("http://127.0.0.1:6600/cart"),
		// UIMode:             stripe.String("embedded"),
		CustomerCreation: stripe.String("always"),
	}
	return params
}

// Runs after the order completes
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔔 Webhook received")

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
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
		var checkoutSession stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &checkoutSession)
		if err != nil {
			http.Error(w, "Failed to parse webhook", http.StatusBadRequest)
			return
		}

		var address *stripe.Address

		if checkoutSession.CollectedInformation.ShippingDetails != nil {
			address = checkoutSession.CollectedInformation.ShippingDetails.Address
			name := checkoutSession.CollectedInformation.ShippingDetails.Name

			fmt.Println("📦 Shipping to:", name)
			fmt.Println("📍 Address:", address.Line1, address.City, address.PostalCode, address.Country)

			// Use this info to store in your DB:
			services.SaveShippingAddress(name, address.Line1, address.City, address.PostalCode, address.Country)
		}

		email := checkoutSession.CustomerDetails.Email
		fmt.Println("📧 Email:", email)

		// Now fetch the full session and expand line_items
		var ev stripe.Event
		if err := json.Unmarshal(payload, &ev); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if ev.Type != "checkout.session.completed" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Unmarshal the minimal session from the event
		var minimalSess stripe.CheckoutSession
		if err := json.Unmarshal(ev.Data.Raw, &minimalSess); err != nil {
			http.Error(w, "could not parse session", http.StatusBadRequest)
			return
		}
		getParams := &stripe.CheckoutSessionParams{}
		getParams.AddExpand("line_items.data.price.product")
		fullSess, err := sessionpkg.Get(minimalSess.ID, getParams)
		if err != nil {
			fmt.Println("❌ Could not fetch full session:", err)
			http.Error(w, "failed to fetch session", http.StatusInternalServerError)
			return
		}

		var items []models.OrderItem
		for _, li := range fullSess.LineItems.Data {
			variantID, _ := strconv.Atoi(li.Price.Product.Metadata["variant_id"])
			items = append(items, models.OrderItem{
				VariantID:    variantID,
				Quantity:     int(li.Quantity),
				Cents:        li.Price.UnitAmount,
				ProductTitle: li.Price.Product.Name,
				VariantColor: li.Price.Product.Description, // or from Metadata if you store color there
			})
		}

		// var items []models.OrderItem
		order_id := services.GenerateShortOrderID()

		// TODO: Match session.ID or customer ID to user/cart
		services.CreateOrder(order_id, email, address.Line1, address.City, address.PostalCode, address.Country, items)
		// and then clear the cart or mark order as paid
		services.ClearCart()

		services.EmailOrderDetails(email)

		fmt.Println("✅ Payment successful for session:", checkoutSession.ID)
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
