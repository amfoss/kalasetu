package app_test

import (
	"errors"
	"sync"
	"testing"

	"kalasetu/testutil"
)

type checkoutSessionItem struct {
	ListingID string  `json:"listingId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type checkoutSessionData struct {
	ID              string                `json:"id"`
	Items           []checkoutSessionItem `json:"items"`
	Total           float64               `json:"total"`
	GatewayOrderID  string                `json:"gatewayOrderId"`
	KeyID           string                `json:"keyId"`
	ExpiresAt       string                `json:"expiresAt"`
	ShippingAddress struct {
		Name       string `json:"name"`
		Line2      string `json:"line2"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	} `json:"shippingAddress"`
}

const checkoutSessionFields = `id total gatewayOrderId keyId expiresAt ` +
	`items { listingId title price quantity } shippingAddress { name line2 postalCode country }`

const createCheckoutSessionMutation = `mutation($a: ShippingAddressInput!) { createCheckoutSession(shippingAddress: $a) { ` + checkoutSessionFields + ` } }`

func checkoutSessionCount(t *testing.T, h *testutil.Harness) int {
	t.Helper()
	var n int
	if err := h.DB.QueryRow(`SELECT count(*) FROM checkout_sessions`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCreateCheckoutSessionReservesStockAndReturnsGatewayDetails(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, pottery, map[string]any{"title": "vase", "price": 100.10, "stock": 5})
	bowl := seedListing(t, h, seller, pottery, map[string]any{"title": "bowl", "price": 33.33, "stock": 2})
	addToCart(t, h, buyer, vase, 3)
	addToCart(t, h, buyer, bowl, 1)

	var data struct {
		CreateCheckoutSession checkoutSessionData `json:"createCheckoutSession"`
	}
	h.GraphQL(t, buyer.Token, createCheckoutSessionMutation, checkoutVars(), &data)
	s := data.CreateCheckoutSession

	if s.Total != 333.63 || len(s.Items) != 2 {
		t.Fatalf("session = %+v, want total 333.63, 2 items", s)
	}
	if s.GatewayOrderID == "" || s.KeyID != "fake_key_id" || s.ExpiresAt == "" {
		t.Errorf("session = %+v, want gateway order id, key id and expiry set", s)
	}
	want := []checkoutSessionItem{
		{ListingID: vase, Title: "vase", Price: 100.10, Quantity: 3},
		{ListingID: bowl, Title: "bowl", Price: 33.33, Quantity: 1},
	}
	for i, w := range want {
		if s.Items[i] != w {
			t.Errorf("item %d = %+v, want %+v", i, s.Items[i], w)
		}
	}
	if s.ShippingAddress.Name != "Asha Rao" || s.ShippingAddress.PostalCode != "682001" || s.ShippingAddress.Line2 != "" {
		t.Errorf("shipping = %+v", s.ShippingAddress)
	}
	if got := stockOf(t, h, vase); got != 2 {
		t.Errorf("vase stock = %d, want 2", got)
	}
	if got := stockOf(t, h, bowl); got != 1 {
		t.Errorf("bowl stock = %d, want 1", got)
	}
	// The Cart is not emptied by session creation.
	if cart := myCart(t, h, buyer); len(cart.Lines) != 2 {
		t.Errorf("cart = %+v, want the 2 lines still there", cart.Lines)
	}
	orders := h.Gateway.Orders()
	if len(orders) != 1 || orders[0].Amount != 33363 || orders[0].Currency != "INR" {
		t.Errorf("gateway orders = %+v, want one of 33363 INR", orders)
	}
}

func TestCreateCheckoutSessionRequiresAuthenticationAndNonEmptyCart(t *testing.T) {
	h := testutil.NewHarness(t)
	buyer := h.CreateUser(t, "Audience")

	requireError(t, h.GraphQLRaw(t, "", createCheckoutSessionMutation, checkoutVars()), "authentication required")
	requireError(t, h.GraphQLRaw(t, buyer.Token, createCheckoutSessionMutation, checkoutVars()), "cart is empty")
	if n := len(h.Gateway.Orders()); n != 0 {
		t.Errorf("gateway orders = %d, want 0", n)
	}
}

func TestCreateCheckoutSessionRequiresCompleteShippingAddress(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)

	blank := address()
	blank["city"] = "   "
	requireError(t, h.GraphQLRaw(t, buyer.Token, createCheckoutSessionMutation, map[string]any{"a": blank}), "shipping address is incomplete")

	if stockOf(t, h, vase) != 5 || checkoutSessionCount(t, h) != 0 {
		t.Error("rejected session changed stock or was recorded")
	}
}

func TestCreateCheckoutSessionFailsWholeWhenAnyLineIsUnavailable(t *testing.T) {
	cases := map[string]func(t *testing.T, h *testutil.Harness, listing string){
		"insufficient stock": func(t *testing.T, h *testutil.Harness, listing string) {
			if _, err := h.DB.Exec(`UPDATE listings SET stock = 1 WHERE id = $1`, listing); err != nil {
				t.Fatal(err)
			}
		},
		"archived": func(t *testing.T, h *testutil.Harness, listing string) { archive(t, h, listing) },
	}
	for name, breakListing := range cases {
		t.Run(name, func(t *testing.T) {
			h := testutil.NewHarness(t)
			seller := h.CreateUser(t, "Artist")
			buyer := h.CreateUser(t, "Audience")
			pottery := categoryID(t, h, "Pottery")
			fine := seedListing(t, h, seller, pottery, map[string]any{"title": "fine", "stock": 5})
			bad := seedListing(t, h, seller, pottery, map[string]any{"title": "the-offender", "stock": 5})
			addToCart(t, h, buyer, fine, 2)
			addToCart(t, h, buyer, bad, 2)
			breakListing(t, h, bad)
			badStock := stockOf(t, h, bad)

			requireError(t, h.GraphQLRaw(t, buyer.Token, createCheckoutSessionMutation, checkoutVars()), "the-offender")

			if got := stockOf(t, h, fine); got != 5 {
				t.Errorf("fine stock = %d, want 5 (untouched)", got)
			}
			if got := stockOf(t, h, bad); got != badStock {
				t.Errorf("offender stock = %d, want %d", got, badStock)
			}
			if n := checkoutSessionCount(t, h); n != 0 {
				t.Errorf("checkout sessions = %d, want 0", n)
			}
			if n := len(h.Gateway.Orders()); n != 0 {
				t.Errorf("gateway orders = %d, want 0", n)
			}
		})
	}
}

func TestCreateCheckoutSessionGatewayFailureReservesNothing(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)
	h.Gateway.FailCreateOrder(errors.New("gateway unreachable"))

	requireError(t, h.GraphQLRaw(t, buyer.Token, createCheckoutSessionMutation, checkoutVars()), "gateway unreachable")

	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5", got)
	}
	if n := checkoutSessionCount(t, h); n != 0 {
		t.Errorf("checkout sessions = %d, want 0", n)
	}
	if cart := myCart(t, h, buyer); len(cart.Lines) != 1 || cart.Lines[0].Quantity != 2 {
		t.Errorf("cart = %+v, want intact", cart.Lines)
	}

	// The buyer can simply retry once the gateway works again.
	h.Gateway.FailCreateOrder(nil)
	var data struct {
		CreateCheckoutSession checkoutSessionData `json:"createCheckoutSession"`
	}
	h.GraphQL(t, buyer.Token, createCheckoutSessionMutation, checkoutVars(), &data)
	if data.CreateCheckoutSession.GatewayOrderID == "" {
		t.Error("retry did not produce a gateway order id")
	}
}

func TestCreateCheckoutSessionRejectsWhenBuyerAlreadyHasOne(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	first := seedListing(t, h, seller, pottery, map[string]any{"title": "first", "stock": 5})
	second := seedListing(t, h, seller, pottery, map[string]any{"title": "second", "stock": 5})
	addToCart(t, h, buyer, first, 1)
	addToCart(t, h, buyer, second, 1)

	var data struct {
		CreateCheckoutSession checkoutSessionData `json:"createCheckoutSession"`
	}
	h.GraphQL(t, buyer.Token, createCheckoutSessionMutation, checkoutVars(), &data)

	requireError(t, h.GraphQLRaw(t, buyer.Token, createCheckoutSessionMutation, checkoutVars()), "already have a checkout in progress")

	// Only the first session's line was reserved.
	if got := stockOf(t, h, first); got != 4 {
		t.Errorf("first stock = %d, want 4", got)
	}
	if got := stockOf(t, h, second); got != 4 {
		t.Errorf("second stock = %d, want 4", got)
	}
	if n := checkoutSessionCount(t, h); n != 1 {
		t.Errorf("checkout sessions = %d, want 1", n)
	}
}

func TestCreateCheckoutSessionAgainstUnconfiguredGatewayFailsClearly(t *testing.T) {
	h := testutil.NewHarnessUnconfiguredGateway(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)

	requireError(t, h.GraphQLRaw(t, buyer.Token, createCheckoutSessionMutation, checkoutVars()), "not configured")

	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5", got)
	}
}

func TestTwoBuyersRacingForLastUnitCheckoutSessionExactlyOneWins(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	piece := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "one-off", "stock": 1})
	const buyers = 8
	users := make([]testutil.User, buyers)
	for i := range users {
		users[i] = h.CreateUser(t, "Audience")
		addToCart(t, h, users[i], piece, 1)
	}

	results := make([]testutil.GraphQLResponse, buyers)
	var wg sync.WaitGroup
	for i := range users {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = h.GraphQLRaw(t, users[i].Token, createCheckoutSessionMutation, checkoutVars())
		}()
	}
	wg.Wait()

	wins := 0
	for _, res := range results {
		if len(res.Errors) == 0 {
			wins++
		} else {
			requireError(t, res, "one-off")
		}
	}
	if wins != 1 {
		t.Fatalf("successful sessions = %d, want exactly 1", wins)
	}
	if got := stockOf(t, h, piece); got != 0 {
		t.Errorf("stock = %d, want 0", got)
	}
	if n := len(h.Gateway.Orders()); n != 1 {
		t.Errorf("gateway orders = %d, want 1", n)
	}
}
