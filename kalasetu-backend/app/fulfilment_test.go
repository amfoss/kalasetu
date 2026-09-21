package app_test

import (
	"testing"

	"kalasetu/testutil"
)

type sellerOrderItem struct {
	ID              string `json:"id"`
	OrderID         string `json:"orderId"`
	ListingID       string `json:"listingId"`
	Title           string `json:"title"`
	Quantity        int    `json:"quantity"`
	Status          string `json:"status"`
	ShippingAddress struct {
		Name       string `json:"name"`
		Line1      string `json:"line1"`
		PostalCode string `json:"postalCode"`
	} `json:"shippingAddress"`
}

const (
	sellerItemFields = `id orderId listingId title quantity status shippingAddress { name line1 postalCode }`
	sellerItemsQuery = `query($s: FulfilmentStatus) { sellerOrderItems(status: $s) { ` + sellerItemFields + ` } }`
	updateItemMut    = `mutation($id: ID!, $s: FulfilmentStatus!) { updateOrderItemStatus(id: $id, status: $s) { ` + sellerItemFields + ` } }`
)

func sellerItems(t *testing.T, h *testutil.Harness, token string, status any) []sellerOrderItem {
	t.Helper()
	var data struct {
		SellerOrderItems []sellerOrderItem `json:"sellerOrderItems"`
	}
	h.GraphQL(t, token, sellerItemsQuery, map[string]any{"s": status}, &data)
	return data.SellerOrderItems
}

// buy makes the buyer check out one unit of each listing in a single Order.
func buy(t *testing.T, h *testutil.Harness, buyer testutil.User, listings ...string) orderData {
	t.Helper()
	for _, l := range listings {
		addToCart(t, h, buyer, l, 1)
	}
	var data struct {
		Checkout orderData `json:"checkout"`
	}
	h.GraphQL(t, buyer.Token, checkoutMutation, checkoutVars(), &data)
	return data.Checkout
}

func TestSellerSeesOnlyOwnOrderItemsWithShippingAddress(t *testing.T) {
	h := testutil.NewHarness(t)
	sellerA := h.CreateUser(t, "Artist")
	sellerB := h.CreateUser(t, "Craftsperson")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, sellerA, pottery, map[string]any{"title": "vase", "stock": 5})
	bowl := seedListing(t, h, sellerB, pottery, map[string]any{"title": "bowl", "stock": 5})
	buy(t, h, buyer, vase, bowl)

	items := sellerItems(t, h, sellerA.Token, nil)
	if len(items) != 1 || items[0].Title != "vase" || items[0].ListingID != vase || items[0].Status != "PAID" {
		t.Fatalf("seller A items = %+v, want just the vase, PAID", items)
	}
	if a := items[0].ShippingAddress; a.Name != "Asha Rao" || a.Line1 != "12 Temple Rd" || a.PostalCode != "682001" {
		t.Errorf("shipping = %+v", a)
	}
	if items := sellerItems(t, h, sellerB.Token, nil); len(items) != 1 || items[0].Title != "bowl" {
		t.Errorf("seller B items = %+v, want just the bowl", items)
	}
	if items := sellerItems(t, h, buyer.Token, nil); len(items) != 0 {
		t.Errorf("buyer items = %+v, want none", items)
	}
	requireError(t, h.GraphQLRaw(t, "", sellerItemsQuery, nil), "authentication required")
}

func updateItem(t *testing.T, h *testutil.Harness, token, id, status string) testutil.GraphQLResponse {
	t.Helper()
	return h.GraphQLRaw(t, token, updateItemMut, map[string]any{"id": id, "s": status})
}

func TestSellerAdvancesItemAndBuyerSeesItInMyOrders(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, pottery, map[string]any{"title": "vase", "stock": 5})
	bowl := seedListing(t, h, seller, pottery, map[string]any{"title": "bowl", "stock": 5})
	buy(t, h, buyer, vase, bowl)
	items := sellerItems(t, h, seller.Token, nil)
	byTitle := map[string]string{}
	for _, it := range items {
		byTitle[it.Title] = it.ID
	}

	var data struct {
		UpdateOrderItemStatus sellerOrderItem `json:"updateOrderItemStatus"`
	}
	h.GraphQL(t, seller.Token, updateItemMut, map[string]any{"id": byTitle["vase"], "s": "SHIPPED"}, &data)
	if got := data.UpdateOrderItemStatus; got.Status != "SHIPPED" || got.Title != "vase" {
		t.Fatalf("updated = %+v, want vase SHIPPED", got)
	}

	if got := sellerItems(t, h, seller.Token, "SHIPPED"); len(got) != 1 || got[0].Title != "vase" {
		t.Errorf("SHIPPED filter = %+v, want just the vase", got)
	}
	if got := sellerItems(t, h, seller.Token, "PAID"); len(got) != 1 || got[0].Title != "bowl" {
		t.Errorf("PAID filter = %+v, want just the bowl", got)
	}
	if got := sellerItems(t, h, seller.Token, "DELIVERED"); len(got) != 0 {
		t.Errorf("DELIVERED filter = %+v, want none", got)
	}

	var orders struct {
		MyOrders []orderData `json:"myOrders"`
	}
	h.GraphQL(t, buyer.Token, myOrdersQuery, nil, &orders)
	o := orders.MyOrders[0]
	if o.Items[0].Status != "SHIPPED" || o.Items[1].Status != "PAID" || o.State != "PARTIALLY_SHIPPED" {
		t.Errorf("buyer sees %+v state %s, want vase SHIPPED, bowl PAID, PARTIALLY_SHIPPED", o.Items, o.State)
	}
}

func TestOnlyTheItemsSellerMayUpdateIt(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	other := h.CreateUser(t, "Craftsperson")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	buy(t, h, buyer, vase)
	id := sellerItems(t, h, seller.Token, nil)[0].ID

	requireError(t, updateItem(t, h, "", id, "SHIPPED"), "authentication required")
	requireError(t, updateItem(t, h, other.Token, id, "SHIPPED"), "forbidden")
	requireError(t, updateItem(t, h, buyer.Token, id, "SHIPPED"), "forbidden")
	requireError(t, updateItem(t, h, seller.Token, "999999", "SHIPPED"), "not found")
	requireError(t, updateItem(t, h, seller.Token, "abc", "SHIPPED"), "invalid")

	if got := sellerItems(t, h, seller.Token, nil)[0].Status; got != "PAID" {
		t.Errorf("status = %s after rejected updates, want PAID", got)
	}
}

func TestOrderItemStatusMovesForwardOneStepAtATime(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	buy(t, h, buyer, vase)
	id := sellerItems(t, h, seller.Token, nil)[0].ID

	statusNow := func() string { return sellerItems(t, h, seller.Token, nil)[0].Status }
	rejected := func(to, from string) {
		t.Helper()
		requireError(t, updateItem(t, h, seller.Token, id, to), "invalid status transition")
		if got := statusNow(); got != from {
			t.Errorf("after rejected %s -> %s status = %s, want unchanged", from, to, got)
		}
	}
	accepted := func(to string) {
		t.Helper()
		res := updateItem(t, h, seller.Token, id, to)
		if len(res.Errors) != 0 {
			t.Fatalf("-> %s: unexpected errors %+v", to, res.Errors)
		}
		if got := statusNow(); got != to {
			t.Fatalf("status = %s, want %s", got, to)
		}
	}

	// From PAID: only SHIPPED is allowed.
	for _, to := range []string{"PAID", "DELIVERED", "CANCELLED"} {
		rejected(to, "PAID")
	}
	accepted("SHIPPED")

	// From SHIPPED: only DELIVERED is allowed.
	for _, to := range []string{"PAID", "SHIPPED", "CANCELLED"} {
		rejected(to, "SHIPPED")
	}
	accepted("DELIVERED")

	// DELIVERED is final.
	for _, to := range []string{"PAID", "SHIPPED", "DELIVERED", "CANCELLED"} {
		rejected(to, "DELIVERED")
	}
}
