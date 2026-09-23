package app_test

import (
	"fmt"
	"net/http"
	"testing"

	"kalasetu/testutil"
)

// webhookBodyWithAmount builds a raw Razorpay-shaped payment webhook payload
// like webhookBody, but also carrying the payment's amount (paise) - the
// only account of a payment's size a webhook backstop has for a gateway
// order id that matches no Checkout Session at all.
func webhookBodyWithAmount(event, gatewayOrderID, paymentID string, amount int64) []byte {
	return []byte(fmt.Sprintf(
		`{"event":%q,"payload":{"payment":{"entity":{"id":%q,"order_id":%q,"amount":%d}}}}`,
		event, paymentID, gatewayOrderID, amount))
}

// checkoutSessionRefundState reads the refund identifier and status
// persisted directly on a Checkout Session, the way refundStateOf reads them
// for an Order Item.
func checkoutSessionRefundState(t *testing.T, h *testutil.Harness, sessionID string) (refundID, refundStatus string) {
	t.Helper()
	if err := h.DB.QueryRow(`SELECT coalesce(refund_id, ''), coalesce(refund_status, '') FROM checkout_sessions WHERE id = $1`, sessionID).
		Scan(&refundID, &refundStatus); err != nil {
		t.Fatal(err)
	}
	return refundID, refundStatus
}

// TestLatePaymentAfterExpiryWithStockGoneRefundsAutomaticallyAndCreatesNoOrder
// covers genuine contention: by the time a late payment arrives, another
// Buyer has already taken the only unit. Fulfilment cannot honour it, so it
// refunds the payment automatically instead of leaving the Buyer to chase
// KalaSetu, and no Order is created for it.
func TestLatePaymentAfterExpiryWithStockGoneRefundsAutomaticallyAndCreatesNoOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	other := h.CreateUser(t, "Audience")
	piece := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "one-off", "stock": 1})
	// Both buyers add to Cart while Stock is still available; adding to Cart
	// checks current availability, so the second buyer's line must be in
	// place before the first buyer's session reserves the only unit.
	addToCart(t, h, buyer, piece, 1)
	addToCart(t, h, other, piece, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)
	expireSession(t, h, session.ID)

	// Someone else buys the only unit while the first buyer's payment is
	// still in flight; opening their session also releases the abandoned one.
	othersOrder := checkout(t, h, other)
	if othersOrder.State != "PAID" {
		t.Fatalf("other buyer's order state = %s, want PAID", othersOrder.State)
	}

	requireError(t, confirmPayment(t, h, buyer.Token, confirmation), "refunded automatically")

	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want 1 (only the other buyer's)", n)
	}
	if got := stockOf(t, h, piece); got != 0 {
		t.Errorf("stock = %d, want 0 (held by the other buyer's order)", got)
	}
	refunds := h.Gateway.Refunds()
	if len(refunds) != 1 || refunds[0].PaymentID != confirmation.PaymentID {
		t.Fatalf("refunds = %+v, want exactly one for the late payment", refunds)
	}
	refundID, refundStatus := checkoutSessionRefundState(t, h, session.ID)
	if refundID == "" || refundStatus != "accepted" {
		t.Errorf("refund_id=%q refund_status=%q, want a refund id and status accepted", refundID, refundStatus)
	}

	// A retried confirmation (e.g. the Buyer's client retrying after a
	// timeout) must not refund a second time: the refund carries a stable
	// idempotency key, the same discipline a cancellation refund uses.
	requireError(t, confirmPayment(t, h, buyer.Token, confirmation), "refunded automatically")
	if n := len(h.Gateway.Refunds()); n != 1 {
		t.Errorf("refunds after retry = %d, want still 1", n)
	}

	// Even if Stock frees up again later (e.g. the other buyer's item is
	// cancelled), a session already refunded must stay refunded: fulfilling
	// it now would ship an Order the Buyer was already made whole for.
	otherItemID := sellerItems(t, h, seller.Token, nil)[0].ID
	if res := cancelItem(t, h, other.Token, otherItemID); len(res.Errors) != 0 {
		t.Fatalf("cancel other buyer's item: %+v", res.Errors)
	}
	requireError(t, confirmPayment(t, h, buyer.Token, confirmation), "refunded automatically")
	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want still 1 (a refunded late payment must not be fulfilled later)", n)
	}
	refundsForLatePayment := 0
	for _, r := range h.Gateway.Refunds() {
		if r.PaymentID == confirmation.PaymentID {
			refundsForLatePayment++
		}
	}
	if refundsForLatePayment != 1 {
		t.Errorf("refunds for the late payment = %d, want still 1 (no re-fulfilment, no second refund)", refundsForLatePayment)
	}
}

// TestWebhookPaymentWithNoMatchingSessionIsRecordedAndRefunded covers a
// payment the gateway reports for a gateway order id no Checkout Session was
// ever opened for: it is recorded rather than ignored, and refunded
// automatically.
func TestWebhookPaymentWithNoMatchingSessionIsRecordedAndRefunded(t *testing.T) {
	h := testutil.NewHarness(t)

	body := webhookBodyWithAmount("payment.captured", "order_does_not_exist", "pay_orphan", 4999)
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if n := orderCount(t, h); n != 0 {
		t.Errorf("orders = %d, want 0 (nothing to fulfil)", n)
	}
	refunds := h.Gateway.Refunds()
	if len(refunds) != 1 || refunds[0].PaymentID != "pay_orphan" || refunds[0].Amount != 4999 {
		t.Fatalf("refunds = %+v, want one refund of 4999 for pay_orphan", refunds)
	}

	var n int
	var reason, refundID, refundStatus string
	if err := h.DB.QueryRow(`
		SELECT count(*), max(reason), max(coalesce(refund_id, '')), max(coalesce(refund_status, ''))
		FROM unmatched_payments WHERE payment_id = $1`, "pay_orphan").
		Scan(&n, &reason, &refundID, &refundStatus); err != nil {
		t.Fatal(err)
	}
	if n != 1 || reason == "" || refundID == "" || refundStatus != "accepted" {
		t.Errorf("unmatched_payments row: count=%d reason=%q refund_id=%q refund_status=%q, want 1 row fully recorded",
			n, reason, refundID, refundStatus)
	}

	// A redelivery of the same notification must not refund twice.
	rec = postWebhook(t, h, body, h.Gateway.SignWebhook(body))
	if rec.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if n := len(h.Gateway.Refunds()); n != 1 {
		t.Errorf("refunds after replay = %d, want still 1", n)
	}
}
