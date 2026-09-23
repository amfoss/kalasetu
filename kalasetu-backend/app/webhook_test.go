package app_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

// refundWebhookBody builds a raw Razorpay-shaped webhook payload for event
// (refund.processed or refund.failed) carrying refundID.
func refundWebhookBody(event, refundID string) []byte {
	return []byte(fmt.Sprintf(
		`{"event":%q,"payload":{"refund":{"entity":{"id":%q}}}}`,
		event, refundID))
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

// webhookEventCount reports how many webhook_events rows are recorded for
// the event identifier Handle derives from body (a sha256 of the raw bytes),
// so a test can check a redelivered notification was recognised rather than
// processed twice.
func webhookEventCount(t *testing.T, h *testutil.Harness, body []byte) int {
	t.Helper()
	sum := sha256.Sum256(body)
	eventID := hex.EncodeToString(sum[:])
	var n int
	if err := h.DB.QueryRow(`SELECT count(*) FROM webhook_events WHERE event_id = $1`, eventID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// cancelledItemWithRefund checks out and cancels one unit of vase, returning
// the cancelled Order Item's id and the gateway refund identifier recorded
// on it.
func cancelledItemWithRefund(t *testing.T, h *testutil.Harness, seller, buyer testutil.User, vase string) (id, refundID string) {
	t.Helper()
	id = buyQty(t, h, buyer, seller, vase, 1)
	if res := cancelItem(t, h, buyer.Token, id); len(res.Errors) != 0 {
		t.Fatalf("cancel: %+v", res.Errors)
	}
	refundID, refundStatus := refundStateOf(t, h, id)
	if refundID == "" || refundStatus != "accepted" {
		t.Fatalf("refund_id %q refund_status %q, want a refund id and status accepted", refundID, refundStatus)
	}
	return id, refundID
}

func TestRefundProcessedWebhookSettlesTheRefund(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id, refundID := cancelledItemWithRefund(t, h, seller, buyer, vase)

	body := refundWebhookBody("refund.processed", refundID)
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if _, status := refundStateOf(t, h, id); status != "settled" {
		t.Errorf("refund_status = %s, want settled", status)
	}
	if got := sellerItems(t, h, seller.Token, nil)[0].Status; got != "CANCELLED" {
		t.Errorf("fulfilment status = %s, want CANCELLED unchanged by refund settlement", got)
	}
}

func TestRefundFailedWebhookFailsTheRefund(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id, refundID := cancelledItemWithRefund(t, h, seller, buyer, vase)

	body := refundWebhookBody("refund.failed", refundID)
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if _, status := refundStateOf(t, h, id); status != "failed" {
		t.Errorf("refund_status = %s, want failed", status)
	}
	if got := sellerItems(t, h, seller.Token, nil)[0].Status; got != "CANCELLED" {
		t.Errorf("fulfilment status = %s, want CANCELLED unchanged by a failed refund", got)
	}
}

func TestRefundWebhookReplayedEventIsANoOp(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id, refundID := cancelledItemWithRefund(t, h, seller, buyer, vase)

	body := refundWebhookBody("refund.processed", refundID)
	signature := h.Gateway.SignWebhook(body)

	first := postWebhook(t, h, body, signature)
	second := postWebhook(t, h, body, signature)

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("status = %d, %d, want 200, 200", first.Code, second.Code)
	}
	if _, status := refundStateOf(t, h, id); status != "settled" {
		t.Errorf("refund_status = %s, want settled", status)
	}
	if n := webhookEventCount(t, h, body); n != 1 {
		t.Errorf("webhook_events rows for this event = %d, want 1 (replay recognised, not reprocessed)", n)
	}
}

func TestRefundWebhookOutOfOrderDeliveryDoesNotFlipASettledOutcome(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id, refundID := cancelledItemWithRefund(t, h, seller, buyer, vase)

	settled := refundWebhookBody("refund.processed", refundID)
	if rec := postWebhook(t, h, settled, h.Gateway.SignWebhook(settled)); rec.Code != http.StatusOK {
		t.Fatalf("settle: status = %d, want 200: %s", rec.Code, rec.Body)
	}

	// A refund.failed for the same refund arrives afterwards (e.g.
	// redelivered out of order): it must not undo the settlement already
	// recorded.
	failed := refundWebhookBody("refund.failed", refundID)
	rec := postWebhook(t, h, failed, h.Gateway.SignWebhook(failed))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if _, status := refundStateOf(t, h, id); status != "settled" {
		t.Errorf("refund_status = %s, want settled (unchanged by the out-of-order refund.failed)", status)
	}
}

func TestRefundWebhookForUnknownRefundIDIsAcknowledgedWithoutError(t *testing.T) {
	h := testutil.NewHarness(t)

	body := refundWebhookBody("refund.processed", "rfnd_unknown")
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
}
