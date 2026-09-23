package app_test

import (
	"testing"

	"kalasetu/testutil"
)

const cancelCheckoutSessionMutation = `mutation { cancelCheckoutSession }`

// expireSession forces session's expiry into the past, the way the harness's
// database handle lets a test simulate an abandoned checkout without any
// clock abstraction or running timer.
func expireSession(t *testing.T, h *testutil.Harness, sessionID string) {
	t.Helper()
	if _, err := h.DB.Exec(`UPDATE checkout_sessions SET expires_at = now() - interval '1 minute' WHERE id = $1`, sessionID); err != nil {
		t.Fatal(err)
	}
}

func sessionStatus(t *testing.T, h *testutil.Harness, sessionID string) string {
	t.Helper()
	var status string
	if err := h.DB.QueryRow(`SELECT status FROM checkout_sessions WHERE id = $1`, sessionID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	return status
}

func TestExpiredCheckoutSessionIsReleasedLazilyBeforeANewOneIsCreated(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)
	first := openCheckoutSession(t, h, buyer)
	if got := stockOf(t, h, vase); got != 3 {
		t.Fatalf("stock after first session = %d, want 3", got)
	}
	expireSession(t, h, first.ID)

	// A second attempt is not rejected as "already have a checkout in
	// progress": the expired one is released first, lazily, before the
	// contended Stock update the new session needs.
	second := openCheckoutSession(t, h, buyer)

	if got := sessionStatus(t, h, first.ID); got != "expired" {
		t.Errorf("first session status = %s, want expired", got)
	}
	if second.ID == first.ID {
		t.Fatal("expected a new session, not the expired one")
	}
	if got := stockOf(t, h, vase); got != 3 {
		t.Errorf("stock after second session = %d, want 3 (restored then reserved once)", got)
	}
}

func TestExpiredCheckoutSessionIsReleasedLazilyForAnotherBuyer(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	first := h.CreateUser(t, "Audience")
	second := h.CreateUser(t, "Audience")
	piece := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "one-off", "stock": 1})
	// Both Buyers add to Cart while Stock is still available; adding to Cart
	// checks current availability, so the second Buyer's line must be in
	// place before the first Buyer's session reserves the only unit.
	addToCart(t, h, first, piece, 1)
	addToCart(t, h, second, piece, 1)
	abandoned := openCheckoutSession(t, h, first)
	expireSession(t, h, abandoned.ID)

	// The Listing looks unavailable while the abandoned session is still
	// "open" in the database, but a second Buyer's attempt must not be
	// turned away for it: the expired session is released lazily right
	// before the contended Stock update this attempt needs, not just the
	// acting Buyer's own.
	secondSession := openCheckoutSession(t, h, second)

	if got := sessionStatus(t, h, abandoned.ID); got != "expired" {
		t.Errorf("abandoned session status = %s, want expired", got)
	}
	if secondSession.ID == abandoned.ID {
		t.Fatal("expected a new session for the second buyer")
	}
	if got := stockOf(t, h, piece); got != 0 {
		t.Errorf("stock = %d, want 0 (now reserved by the second buyer)", got)
	}
}

// TestLatePaymentAfterExpiryWithStockAvailableProducesAnOrder covers a
// payment that lands just after its reservation expired, while nobody else
// wanted the item: fulfilment re-acquires the same Stock and honours it,
// exactly as issue 08 requires.
func TestLatePaymentAfterExpiryWithStockAvailableProducesAnOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)
	expireSession(t, h, session.ID)

	res := confirmPayment(t, h, buyer.Token, confirmation)
	if len(res.Errors) != 0 {
		t.Fatalf("late confirm: %+v", res.Errors)
	}

	if orderCount(t, h) != 1 {
		t.Error("late payment with stock available did not produce an order")
	}
	if got := stockOf(t, h, vase); got != 4 {
		t.Errorf("stock = %d, want 4 (re-acquired and consumed by the late payment)", got)
	}
	if got := sessionStatus(t, h, session.ID); got != "consumed" {
		t.Errorf("session status = %s, want consumed", got)
	}
	if n := len(h.Gateway.Refunds()); n != 0 {
		t.Errorf("refunds = %d, want 0 (the late payment was honoured, not refunded)", n)
	}
}

func TestCancelCheckoutSessionReleasesStockAndAllowsANewOne(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)
	session := openCheckoutSession(t, h, buyer)

	var data struct {
		CancelCheckoutSession bool `json:"cancelCheckoutSession"`
	}
	h.GraphQL(t, buyer.Token, cancelCheckoutSessionMutation, nil, &data)
	if !data.CancelCheckoutSession {
		t.Fatal("cancelCheckoutSession = false, want true")
	}

	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5 (released immediately)", got)
	}
	if got := sessionStatus(t, h, session.ID); got != "cancelled" {
		t.Errorf("session status = %s, want cancelled", got)
	}

	// The Buyer is not left waiting out the reservation window: a new
	// Checkout Session can be opened immediately.
	second := openCheckoutSession(t, h, buyer)
	if second.ID == session.ID {
		t.Fatal("expected a new session")
	}
	if got := stockOf(t, h, vase); got != 3 {
		t.Errorf("stock after new session = %d, want 3", got)
	}
}

func TestCancelCheckoutSessionRequiresAnOpenSession(t *testing.T) {
	h := testutil.NewHarness(t)
	buyer := h.CreateUser(t, "Audience")

	requireError(t, h.GraphQLRaw(t, buyer.Token, cancelCheckoutSessionMutation, nil), "no checkout in progress")
}

func TestCancelCheckoutSessionRequiresAuthentication(t *testing.T) {
	h := testutil.NewHarness(t)

	requireError(t, h.GraphQLRaw(t, "", cancelCheckoutSessionMutation, nil), "authentication required")
}

func TestPaymentFailedWebhookReleasesTheCheckoutSessionImmediately(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)
	session := openCheckoutSession(t, h, buyer)

	body := webhookBody("payment.failed", session.GatewayOrderID, "pay_test_failed")
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5 (released immediately, not left for expiry)", got)
	}
	if got := sessionStatus(t, h, session.ID); got != "cancelled" {
		t.Errorf("session status = %s, want cancelled", got)
	}

	// The Buyer can start a new Checkout Session right away.
	second := openCheckoutSession(t, h, buyer)
	if second.ID == session.ID {
		t.Fatal("expected a new session")
	}
}

func TestPaymentFailedWebhookReplayedEventIsANoOp(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)

	body := webhookBody("payment.failed", session.GatewayOrderID, "pay_test_failed")
	signature := h.Gateway.SignWebhook(body)

	first := postWebhook(t, h, body, signature)
	second := postWebhook(t, h, body, signature)

	if first.Code != 200 || second.Code != 200 {
		t.Fatalf("status = %d, %d, want 200, 200", first.Code, second.Code)
	}
	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5 (released exactly once)", got)
	}
}

func TestPaymentFailedWebhookRacingASuccessfulConfirmationDoesNotUndoTheOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	// The Buyer's browser confirms first...
	if res := confirmPayment(t, h, buyer.Token, confirmation); len(res.Errors) != 0 {
		t.Fatalf("confirm: %+v", res.Errors)
	}

	// ...and a stale payment-failed notification for the same order still
	// arrives afterwards.
	body := webhookBody("payment.failed", session.GatewayOrderID, confirmation.PaymentID)
	rec := postWebhook(t, h, body, h.Gateway.SignWebhook(body))

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	if orderCount(t, h) != 1 {
		t.Error("the already-fulfilled order was affected by the stale payment-failed webhook")
	}
	if got := stockOf(t, h, vase); got != 4 {
		t.Errorf("stock = %d, want 4 (still consumed by the paid order)", got)
	}
	if got := sessionStatus(t, h, session.ID); got != "consumed" {
		t.Errorf("session status = %s, want consumed (unaffected by the stale webhook)", got)
	}
}

func TestReleasedListingBecomesBuyableAgain(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	first := h.CreateUser(t, "Audience")
	second := h.CreateUser(t, "Audience")
	piece := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "one-off", "stock": 1})
	addToCart(t, h, first, piece, 1)
	session := openCheckoutSession(t, h, first)
	if got := stockOf(t, h, piece); got != 0 {
		t.Fatalf("stock after reservation = %d, want 0", got)
	}

	// The first Buyer walks away and their session is released explicitly.
	var data struct {
		CancelCheckoutSession bool `json:"cancelCheckoutSession"`
	}
	h.GraphQL(t, first.Token, cancelCheckoutSessionMutation, nil, &data)

	// A second Buyer can now check out the same Listing.
	addToCart(t, h, second, piece, 1)
	placed := checkout(t, h, second)
	if placed.State != "PAID" {
		t.Fatalf("second buyer's order state = %s, want PAID", placed.State)
	}
	if got := stockOf(t, h, piece); got != 0 {
		t.Errorf("stock = %d, want 0 (now held by the second buyer's order)", got)
	}
	if got := sessionStatus(t, h, session.ID); got != "cancelled" {
		t.Errorf("first session status = %s, want cancelled", got)
	}
}
