package app_test

import (
	"testing"

	"kalasetu/payments"
	"kalasetu/testutil"
)

type orderItem struct {
	ListingID string  `json:"listingId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Status    string  `json:"status"`
}

type orderData struct {
	ID              string      `json:"id"`
	Items           []orderItem `json:"items"`
	Total           float64     `json:"total"`
	State           string      `json:"state"`
	ShippingAddress struct {
		Name       string `json:"name"`
		Line2      string `json:"line2"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	} `json:"shippingAddress"`
}

const orderFields = `id total state items { listingId title price quantity status } shippingAddress { name line2 postalCode country }`

const (
	confirmPaymentMutation = `mutation($i: ConfirmCheckoutSessionPaymentInput!) { confirmCheckoutSessionPayment(input: $i) { ` + orderFields + ` } }`
	myOrdersQuery          = `{ myOrders { ` + orderFields + ` } }`
)

func address() map[string]any {
	return map[string]any{
		"name": "Asha Rao", "phone": "9999999999", "line1": "12 Temple Rd",
		"city": "Kochi", "state": "Kerala", "postalCode": "682001", "country": "India",
	}
}

func checkoutVars() map[string]any { return map[string]any{"a": address()} }

func stockOf(t *testing.T, h *testutil.Harness, listingID string) int {
	t.Helper()
	var n int
	if err := h.DB.QueryRow(`SELECT stock FROM listings WHERE id = $1`, listingID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func orderCount(t *testing.T, h *testutil.Harness) int {
	t.Helper()
	var n int
	if err := h.DB.QueryRow(`SELECT count(*) FROM orders`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// openCheckoutSession creates a Checkout Session for the buyer's Cart with
// the default address, failing the test on any error.
func openCheckoutSession(t *testing.T, h *testutil.Harness, buyer testutil.User) checkoutSessionData {
	t.Helper()
	var data struct {
		CreateCheckoutSession checkoutSessionData `json:"createCheckoutSession"`
	}
	h.GraphQL(t, buyer.Token, createCheckoutSessionMutation, checkoutVars(), &data)
	return data.CreateCheckoutSession
}

func confirmVars(c payments.ConfirmationRequest) map[string]any {
	return map[string]any{"i": map[string]any{
		"gatewayOrderId": c.GatewayOrderID, "paymentId": c.PaymentID, "signature": c.Signature,
	}}
}

// confirmPayment relays confirmation to confirmCheckoutSessionPayment, for
// tests that expect an error.
func confirmPayment(t *testing.T, h *testutil.Harness, token string, c payments.ConfirmationRequest) testutil.GraphQLResponse {
	t.Helper()
	return h.GraphQLRaw(t, token, confirmPaymentMutation, confirmVars(c))
}

// checkout opens a Checkout Session for the buyer's Cart and confirms
// payment for it with the harness's fake gateway, returning the Order.
func checkout(t *testing.T, h *testutil.Harness, buyer testutil.User) orderData {
	t.Helper()
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)
	var data struct {
		ConfirmCheckoutSessionPayment orderData `json:"confirmCheckoutSessionPayment"`
	}
	h.GraphQL(t, buyer.Token, confirmPaymentMutation, confirmVars(confirmation), &data)
	return data.ConfirmCheckoutSessionPayment
}

func TestConfirmingPaymentTurnsSessionIntoPaidOrderAndEmptiesCart(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, pottery, map[string]any{"title": "vase", "price": 100.10, "stock": 5})
	bowl := seedListing(t, h, seller, pottery, map[string]any{"title": "bowl", "price": 33.33, "stock": 2})
	addToCart(t, h, buyer, vase, 3)
	addToCart(t, h, buyer, bowl, 1)

	session := openCheckoutSession(t, h, buyer)
	// The Cart stays intact until the payment is confirmed.
	if cart := myCart(t, h, buyer); len(cart.Lines) != 2 {
		t.Fatalf("cart before confirm = %+v, want the 2 lines still there", cart.Lines)
	}

	confirmation := h.Gateway.Confirm(session.GatewayOrderID)
	var data struct {
		ConfirmCheckoutSessionPayment orderData `json:"confirmCheckoutSessionPayment"`
	}
	h.GraphQL(t, buyer.Token, confirmPaymentMutation, confirmVars(confirmation), &data)
	o := data.ConfirmCheckoutSessionPayment

	if o.Total != 333.63 || o.State != "PAID" || len(o.Items) != 2 {
		t.Fatalf("order = %+v, want total 333.63, PAID, 2 items", o)
	}
	want := []orderItem{
		{ListingID: vase, Title: "vase", Price: 100.10, Quantity: 3, Status: "PAID"},
		{ListingID: bowl, Title: "bowl", Price: 33.33, Quantity: 1, Status: "PAID"},
	}
	for i, w := range want {
		if o.Items[i] != w {
			t.Errorf("item %d = %+v, want %+v", i, o.Items[i], w)
		}
	}
	if o.ShippingAddress.Name != "Asha Rao" || o.ShippingAddress.PostalCode != "682001" || o.ShippingAddress.Line2 != "" {
		t.Errorf("shipping = %+v", o.ShippingAddress)
	}
	if got := stockOf(t, h, vase); got != 2 {
		t.Errorf("vase stock = %d, want 2", got)
	}
	if got := stockOf(t, h, bowl); got != 1 {
		t.Errorf("bowl stock = %d, want 1", got)
	}
	if cart := myCart(t, h, buyer); len(cart.Lines) != 0 {
		t.Errorf("cart = %+v, want empty after confirm", cart.Lines)
	}
	var status string
	if err := h.DB.QueryRow(`SELECT status FROM checkout_sessions WHERE id = $1`, session.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "consumed" {
		t.Errorf("session status = %s, want consumed", status)
	}
}

func TestConfirmCheckoutSessionPaymentRequiresAuthentication(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	requireError(t, h.GraphQLRaw(t, "", myOrdersQuery, nil), "authentication required")
	requireError(t, confirmPayment(t, h, "", confirmation), "authentication required")
	if orderCount(t, h) != 0 {
		t.Error("rejected confirm created an order")
	}
}

func TestConfirmCheckoutSessionPaymentRejectsForgedSignature(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)
	confirmation.Signature = "tampered"

	requireError(t, confirmPayment(t, h, buyer.Token, confirmation), "payment confirmation could not be verified")

	if got := stockOf(t, h, vase); got != 4 {
		t.Errorf("stock = %d, want 4 (still reserved, nothing consumed)", got)
	}
	if orderCount(t, h) != 0 {
		t.Error("a forged confirmation produced an order")
	}
	if cart := myCart(t, h, buyer); len(cart.Lines) != 1 {
		t.Errorf("cart = %+v, want intact", cart.Lines)
	}
}

func TestConfirmCheckoutSessionPaymentRejectsWrongBuyer(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	other := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	requireError(t, confirmPayment(t, h, other.Token, confirmation), "forbidden")
	if orderCount(t, h) != 0 {
		t.Error("a confirm by another buyer produced an order")
	}
}

func TestConfirmCheckoutSessionPaymentRejectsUnknownGatewayOrderID(t *testing.T) {
	h := testutil.NewHarness(t)
	buyer := h.CreateUser(t, "Audience")
	confirmation := h.Gateway.Confirm("does-not-exist")

	requireError(t, confirmPayment(t, h, buyer.Token, confirmation), "not found")
}

func TestConfirmCheckoutSessionPaymentRejectsANonOpenSession(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	if _, err := h.DB.Exec(`UPDATE checkout_sessions SET status = 'expired' WHERE id = $1`, session.ID); err != nil {
		t.Fatal(err)
	}

	requireError(t, confirmPayment(t, h, buyer.Token, confirmation), "cannot be confirmed")
	if orderCount(t, h) != 0 {
		t.Error("confirming an expired session produced an order")
	}
}

func TestConfirmingSamePaymentTwiceProducesExactlyOneOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)
	session := openCheckoutSession(t, h, buyer)
	confirmation := h.Gateway.Confirm(session.GatewayOrderID)

	var first, second struct {
		ConfirmCheckoutSessionPayment orderData `json:"confirmCheckoutSessionPayment"`
	}
	h.GraphQL(t, buyer.Token, confirmPaymentMutation, confirmVars(confirmation), &first)
	h.GraphQL(t, buyer.Token, confirmPaymentMutation, confirmVars(confirmation), &second)

	if first.ConfirmCheckoutSessionPayment.ID != second.ConfirmCheckoutSessionPayment.ID {
		t.Errorf("ids = %s, %s, want the same order both times",
			first.ConfirmCheckoutSessionPayment.ID, second.ConfirmCheckoutSessionPayment.ID)
	}
	if n := orderCount(t, h); n != 1 {
		t.Errorf("orders = %d, want 1", n)
	}
	if got := stockOf(t, h, vase); got != 3 {
		t.Errorf("stock = %d, want 3 (decremented once)", got)
	}
}

func TestOrderItemsAreSnapshotsUnaffectedByListingChanges(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "vase", "price": 50, "stock": 3})
	addToCart(t, h, buyer, vase, 1)
	placed := checkout(t, h, buyer)

	if _, err := h.DB.Exec(`UPDATE listings SET title = 'renamed', price = 999 WHERE id = $1`, vase); err != nil {
		t.Fatal(err)
	}
	archive(t, h, vase)

	var orders struct {
		MyOrders []orderData `json:"myOrders"`
	}
	h.GraphQL(t, buyer.Token, myOrdersQuery, nil, &orders)
	if len(orders.MyOrders) != 1 {
		t.Fatalf("orders = %+v, want 1", orders.MyOrders)
	}
	o := orders.MyOrders[0]
	if o.ID != placed.ID || o.Total != 50 || o.State != "PAID" ||
		len(o.Items) != 1 || o.Items[0].Title != "vase" || o.Items[0].Price != 50 || o.Items[0].Status != "PAID" {
		t.Errorf("order = %+v, want the original snapshot", o)
	}
}

func TestMyOrdersReturnsOnlyCallersOrdersNewestFirst(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	other := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	first := seedListing(t, h, seller, pottery, map[string]any{"title": "first", "stock": 2})
	second := seedListing(t, h, seller, pottery, map[string]any{"title": "second", "stock": 2})

	for _, l := range []string{first, second} {
		addToCart(t, h, buyer, l, 1)
		checkout(t, h, buyer)
	}
	addToCart(t, h, other, first, 1)
	checkout(t, h, other)

	var orders struct {
		MyOrders []orderData `json:"myOrders"`
	}
	h.GraphQL(t, buyer.Token, myOrdersQuery, nil, &orders)
	if len(orders.MyOrders) != 2 || orders.MyOrders[0].Items[0].Title != "second" || orders.MyOrders[1].Items[0].Title != "first" {
		t.Fatalf("orders = %+v, want [second, first]", orders.MyOrders)
	}
}

func TestOrderStateIsDerivedFromItemStatuses(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	a := seedListing(t, h, seller, pottery, map[string]any{"stock": 2})
	b := seedListing(t, h, seller, pottery, map[string]any{"stock": 2})
	addToCart(t, h, buyer, a, 1)
	addToCart(t, h, buyer, b, 1)
	checkout(t, h, buyer)

	steps := []struct{ setA, setB, want string }{
		{"shipped", "paid", "PARTIALLY_SHIPPED"},
		{"shipped", "shipped", "SHIPPED"},
		{"delivered", "shipped", "SHIPPED"},
		{"delivered", "delivered", "DELIVERED"},
		{"cancelled", "delivered", "DELIVERED"},
		{"cancelled", "cancelled", "CANCELLED"},
	}
	for _, s := range steps {
		for listing, status := range map[string]string{a: s.setA, b: s.setB} {
			if _, err := h.DB.Exec(`UPDATE order_items SET status = $2 WHERE listing_id = $1`, listing, status); err != nil {
				t.Fatal(err)
			}
		}
		var orders struct {
			MyOrders []orderData `json:"myOrders"`
		}
		h.GraphQL(t, buyer.Token, myOrdersQuery, nil, &orders)
		if got := orders.MyOrders[0].State; got != s.want {
			t.Errorf("items %s/%s: state = %s, want %s", s.setA, s.setB, got, s.want)
		}
	}
}
