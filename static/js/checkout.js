const stripe = Stripe("pk_test_51Rqloa2NPGaRjqznTsWRCdGiNIRGm6R5NsvNoDUx5QmGQamGcZUEUUkFBvZkDDqIiW5G9LrkvAIvO2aHrWjRY0G000dPOEwAPV")

initialize();

// Create a Checkout Session
async function initialize() {
  const fetchClientSecret = async () => {
    const response = await fetch("/cart-checkout", {
      method: "POST",
    });
    const { clientSecret } = await response.json();
    return clientSecret;
  };

  const checkout = await stripe.initEmbeddedCheckout({
    fetchClientSecret,
  });

  // Mount Checkout
  checkout.mount('#checkout');
}