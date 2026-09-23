package app_test

import (
	"database/sql"
	"errors"
	"testing"

	"kalasetu/payments"
	"kalasetu/testutil"
)

const cancelItemMut = `mutation($id: ID!) { cancelOrderItem(id: $id) { ` + sellerItemFields + ` } }`

func cancelItem(t *testing.T, h *testutil.Harness, token, id string) testutil.GraphQLResponse {
	t.Helper()
	return h.GraphQLRaw(t, token, cancelItemMut, map[string]any{"id": id})
}

// buyQty checks out qty units of one listing and returns the seller's view of the new item id.
func buyQty(t *testing.T, h *testutil.Harness, buyer, seller testutil.User, listing string, qty int) string {
	t.Helper()
	addToCart(t, h, buyer, listing, qty)
	checkout(t, h, buyer)
	items := sellerItems(t, h, seller.Token, "PAID")
	return items[0].ID
}

// refundIdempotencyKey mirrors repos.refundIdempotencyKey: a stable key
// derived only from the Order Item's id.
func refundIdempotencyKey(id string) string { return "order-item-refund-" + id }

// refundStateOf reads the refund identifier and status persisted directly on
// the order_items row, bypassing GraphQL, since the ticket only requires the
// state to be written and readable, not exposed over the API yet.
func refundStateOf(t *testing.T, h *testutil.Harness, id string) (refundID, refundStatus string) {
	t.Helper()
	var gotID, gotStatus sql.NullString
	if err := h.DB.QueryRow(`SELECT refund_id, refund_status FROM order_items WHERE id = $1`, id).Scan(&gotID, &gotStatus); err != nil {
		t.Fatal(err)
	}
	return gotID.String, gotStatus.String
}

func TestBuyerCancelsPaidItemRestocksAndRefunds(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5, "price": 12.5})
	id := buyQty(t, h, buyer, seller, vase, 2)
	if got := stockOf(t, h, vase); got != 3 {
		t.Fatalf("stock after checkout = %d, want 3", got)
	}

	var data struct {
		CancelOrderItem sellerOrderItem `json:"cancelOrderItem"`
	}
	h.GraphQL(t, buyer.Token, cancelItemMut, map[string]any{"id": id}, &data)
	if data.CancelOrderItem.Status != "CANCELLED" {
		t.Errorf("status = %s, want CANCELLED", data.CancelOrderItem.Status)
	}
	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5 after restock", got)
	}
	key := refundIdempotencyKey(id)
	refunds := h.Gateway.Refunds()
	want := payments.RefundOrderRequest{PaymentID: "fake_payment_1", Amount: 2500, IdempotencyKey: key, Reference: key}
	if len(refunds) != 1 || refunds[0] != want {
		t.Errorf("refunds = %+v, want one refund %+v", refunds, want)
	}
	refundID, refundStatus := refundStateOf(t, h, id)
	if refundID == "" || refundStatus != "accepted" {
		t.Errorf("refund_id %q refund_status %q, want a refund id and status accepted", refundID, refundStatus)
	}
}

func TestSellerCancelsPaidItem(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id := buyQty(t, h, buyer, seller, vase, 1)

	var data struct {
		CancelOrderItem sellerOrderItem `json:"cancelOrderItem"`
	}
	h.GraphQL(t, seller.Token, cancelItemMut, map[string]any{"id": id}, &data)
	if data.CancelOrderItem.Status != "CANCELLED" || stockOf(t, h, vase) != 5 || len(h.Gateway.Refunds()) != 1 {
		t.Errorf("status %s stock %d refunds %d, want CANCELLED, 5, 1",
			data.CancelOrderItem.Status, stockOf(t, h, vase), len(h.Gateway.Refunds()))
	}
}

func TestOnlyBuyerOrSellerMayCancel(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	other := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id := buyQty(t, h, buyer, seller, vase, 1)

	requireError(t, cancelItem(t, h, "", id), "authentication required")
	requireError(t, cancelItem(t, h, other.Token, id), "forbidden")
	requireError(t, cancelItem(t, h, buyer.Token, "999999"), "not found")
	requireError(t, cancelItem(t, h, buyer.Token, "abc"), "invalid")

	if got := sellerItems(t, h, seller.Token, nil)[0].Status; got != "PAID" || stockOf(t, h, vase) != 4 || len(h.Gateway.Refunds()) != 0 {
		t.Errorf("rejected cancels changed state: status %s stock %d refunds %d", got, stockOf(t, h, vase), len(h.Gateway.Refunds()))
	}
}

func TestOnlyPaidItemsCanBeCancelled(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id := buyQty(t, h, buyer, seller, vase, 1)

	updateItem(t, h, seller.Token, id, "SHIPPED")
	requireError(t, cancelItem(t, h, buyer.Token, id), "cannot be cancelled")
	requireError(t, cancelItem(t, h, seller.Token, id), "cannot be cancelled")
	updateItem(t, h, seller.Token, id, "DELIVERED")
	requireError(t, cancelItem(t, h, buyer.Token, id), "cannot be cancelled")

	if stockOf(t, h, vase) != 4 || len(h.Gateway.Refunds()) != 0 {
		t.Errorf("stock %d refunds %d, want 4 and 0", stockOf(t, h, vase), len(h.Gateway.Refunds()))
	}
}

func TestCancellingTwiceIsRejectedAndRefundsOnce(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id := buyQty(t, h, buyer, seller, vase, 1)

	if res := cancelItem(t, h, buyer.Token, id); len(res.Errors) != 0 {
		t.Fatalf("first cancel: %+v", res.Errors)
	}
	requireError(t, cancelItem(t, h, seller.Token, id), "cannot be cancelled")
	if stockOf(t, h, vase) != 5 || len(h.Gateway.Refunds()) != 1 {
		t.Errorf("stock %d refunds %d, want 5 and 1", stockOf(t, h, vase), len(h.Gateway.Refunds()))
	}
}

func TestCancellingRestocksAnArchivedListing(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id := buyQty(t, h, buyer, seller, vase, 2)
	archive(t, h, vase)

	if res := cancelItem(t, h, buyer.Token, id); len(res.Errors) != 0 {
		t.Fatalf("cancel: %+v", res.Errors)
	}
	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5", got)
	}
}

func TestRefundFailureLeavesItemAndStockUnchanged(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	id := buyQty(t, h, buyer, seller, vase, 2)
	h.Gateway.FailRefund(errors.New("gateway down"))

	requireError(t, cancelItem(t, h, buyer.Token, id), "refund failed")
	if got := sellerItems(t, h, seller.Token, nil)[0].Status; got != "PAID" || stockOf(t, h, vase) != 3 {
		t.Errorf("status %s stock %d, want PAID and 3", got, stockOf(t, h, vase))
	}

	// Retrying once the gateway recovers works.
	h.Gateway.FailRefund(nil)
	if res := cancelItem(t, h, buyer.Token, id); len(res.Errors) != 0 {
		t.Fatalf("retry: %+v", res.Errors)
	}
	if stockOf(t, h, vase) != 5 {
		t.Errorf("stock = %d, want 5", stockOf(t, h, vase))
	}
}

func TestCancellingOneItemLeavesOthersInTheOrderAlone(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, pottery, map[string]any{"title": "vase", "stock": 5, "price": 10.0})
	bowl := seedListing(t, h, seller, pottery, map[string]any{"title": "bowl", "stock": 5, "price": 20.0})
	buy(t, h, buyer, vase, bowl)
	var id string
	for _, it := range sellerItems(t, h, seller.Token, nil) {
		if it.Title == "vase" {
			id = it.ID
		}
	}

	if res := cancelItem(t, h, buyer.Token, id); len(res.Errors) != 0 {
		t.Fatalf("cancel: %+v", res.Errors)
	}
	var orders struct {
		MyOrders []orderData `json:"myOrders"`
	}
	h.GraphQL(t, buyer.Token, myOrdersQuery, nil, &orders)
	o := orders.MyOrders[0]
	if o.Items[0].Status != "CANCELLED" || o.Items[1].Status != "PAID" || o.State != "PAID" {
		t.Errorf("items %+v state %s, want vase CANCELLED, bowl PAID, order PAID", o.Items, o.State)
	}
	if stockOf(t, h, vase) != 5 || stockOf(t, h, bowl) != 4 {
		t.Errorf("stock vase %d bowl %d, want 5 and 4", stockOf(t, h, vase), stockOf(t, h, bowl))
	}
	if r := h.Gateway.Refunds(); len(r) != 1 || r[0].Amount != 1000 {
		t.Errorf("refunds = %+v, want one of 1000", r)
	}
}

func TestCancelRefundAcceptedButPendingStillCommits(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5, "price": 12.5})
	id := buyQty(t, h, buyer, seller, vase, 2)

	// The gateway accepts the refund but has not yet settled it with the bank.
	h.Gateway.NextRefundStatus("pending")

	var data struct {
		CancelOrderItem sellerOrderItem `json:"cancelOrderItem"`
	}
	h.GraphQL(t, buyer.Token, cancelItemMut, map[string]any{"id": id}, &data)
	if data.CancelOrderItem.Status != "CANCELLED" {
		t.Errorf("status = %s, want CANCELLED", data.CancelOrderItem.Status)
	}
	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5 after restock", got)
	}
	refundID, refundStatus := refundStateOf(t, h, id)
	if refundID == "" || refundStatus != "accepted" {
		t.Errorf("refund_id %q refund_status %q, want a refund id and status accepted", refundID, refundStatus)
	}
}

func TestCancelRefundConflictTreatedAsSuccess(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5, "price": 12.5})
	id := buyQty(t, h, buyer, seller, vase, 2)

	// Simulate the gateway already having accepted this refund - the outcome
	// a real gateway would report with a 409 conflict on a repeated
	// idempotency key - before the cancellation ever calls it.
	key := refundIdempotencyKey(id)
	h.Gateway.SeedRefund(key, payments.RefundOrderResult{RefundID: "rfnd_existing", Status: "processed"})

	var data struct {
		CancelOrderItem sellerOrderItem `json:"cancelOrderItem"`
	}
	h.GraphQL(t, buyer.Token, cancelItemMut, map[string]any{"id": id}, &data)
	if data.CancelOrderItem.Status != "CANCELLED" {
		t.Errorf("status = %s, want CANCELLED", data.CancelOrderItem.Status)
	}
	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5 after restock", got)
	}
	// The conflict replay is not a fresh refund.
	if refunds := h.Gateway.Refunds(); len(refunds) != 0 {
		t.Errorf("refunds = %+v, want none: the conflict was a replay, not a fresh refund", refunds)
	}
	refundID, refundStatus := refundStateOf(t, h, id)
	if refundID != "rfnd_existing" || refundStatus != "accepted" {
		t.Errorf("refund_id %q refund_status %q, want rfnd_existing and accepted", refundID, refundStatus)
	}
}
