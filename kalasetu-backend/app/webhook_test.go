package app_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"kalasetu/testutil"
)

// webhookBody builds a raw Razorpay-shaped webhook payload for event
// (payment.captured or order.paid) carrying gatewayOrderID and paymentID.
func webhookBody(event, gatewayOrderID, paymentID string) []byte {
	return []byte(fmt.Sprintf(
		`{"event":%q,"payload":{"payment":{"entity":{"id":%q,"order_id":%q}}}}`,
		event, paymentID, gatewayOrderID))
}

// postWebhook posts body to the Razorpay webhook endpoint signed with
// signature (an empty signature sends no header), returning the raw HTTP
// response.
func postWebhook(t *testing.T, h *testutil.Harness, body []byte, signature string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/razorpay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set("X-Razorpay-Signature", signature)
	}
	rec := httptest.NewRecorder()
	h.App.Router.ServeHTTP(rec, req)
	return rec
}

func TestWebhookFulfilsCheckoutSessionOnValidSignature(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)

	body := webhookBody("payment.captured", session.GatewayOrderID, "pay_test_1")
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want 1", n)
	}
	var status string
	if err := h.DB.QueryRow(`SELECT status FROM checkout_sessions WHERE id = $1`, session.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "consumed" {
		t.Errorf("session status = %s, want consumed", status)
	}
	if cart := myCart(t, h, buyer); len(cart.Lines) != 0 {
		t.Errorf("cart = %+v, want empty after webhook fulfilment", cart.Lines)
	}
}

func TestWebhookRejectsInvalidSignature(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)

	body := webhookBody("payment.captured", session.GatewayOrderID, "pay_test_1")
	rec := postWebhook(t, h, body, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body)
	}
	if n := orderCount(t, h); n != 0 {
		t.Errorf("orders = %d, want 0 (nothing created on an invalid signature)", n)
	}
}

func TestWebhookRejectsAbsentSignature(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)

	body := webhookBody("payment.captured", session.GatewayOrderID, "pay_test_1")
	rec := postWebhook(t, h, body, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body)
	}
	if n := orderCount(t, h); n != 0 {
		t.Errorf("orders = %d, want 0", n)
	}
}

func TestWebhookReplayedEventIsANoOp(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)

	body := webhookBody("payment.captured", session.GatewayOrderID, "pay_test_1")
	signature := h.Gateway.SignWebhook(body)

	first := postWebhook(t, h, body, signature)
	second := postWebhook(t, h, body, signature)

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("status = %d, %d, want 200, 200", first.Code, second.Code)
	}
	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want 1 (replay produced no second order)", n)
	}
}

func TestWebhookRacingConfirmationProducesExactlyOneOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	// The Buyer's browser confirms first...
	var data struct {
		ConfirmCheckoutSessionPayment orderData `json:"confirmCheckoutSessionPayment"`
	}
	h.GraphQL(t, buyer.Token, confirmPaymentMutation, confirmVars(confirmation), &data)

	// ...and Razorpay's webhook for the same payment still arrives.
	body := webhookBody("payment.captured", session.GatewayOrderID, confirmation.PaymentID)
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want exactly 1", n)
	}
	if data.ConfirmCheckoutSessionPayment.ID == "" {
		t.Fatal("confirmation did not return an order")
	}
}

func TestWebhookConfirmationRacingItProducesExactlyOneOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	// Razorpay's webhook (the buyer's browser never reported back) arrives first...
	body := webhookBody("order.paid", session.GatewayOrderID, confirmation.PaymentID)
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}

	// ...and the buyer's browser still confirms afterwards.
	res := confirmPayment(t, h, buyer.Token, confirmation)
	if len(res.Errors) > 0 {
		t.Fatalf("confirm after webhook: %+v", res.Errors)
	}

	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want exactly 1", n)
	}
}
