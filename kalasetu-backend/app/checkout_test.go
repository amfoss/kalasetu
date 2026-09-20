package app_test

import (
	"errors"
	"sync"
	"testing"

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
	checkoutMutation = `mutation($a: ShippingAddressInput!) { checkout(shippingAddress: $a) { ` + orderFields + ` } }`
	myOrdersQuery    = `{ myOrders { ` + orderFields + ` } }`
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

func TestCheckoutTurnsCartIntoPaidOrderAndEmptiesCart(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	pottery := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, pottery, map[string]any{"title": "vase", "price": 100.10, "stock": 5})
	bowl := seedListing(t, h, seller, pottery, map[string]any{"title": "bowl", "price": 33.33, "stock": 2})
	addToCart(t, h, buyer, vase, 3)
	addToCart(t, h, buyer, bowl, 1)

	var data struct {
		Checkout orderData `json:"checkout"`
	}
	h.GraphQL(t, buyer.Token, checkoutMutation, checkoutVars(), &data)
	o := data.Checkout

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
		t.Errorf("cart = %+v, want empty", cart.Lines)
	}
	charges := h.Payments.Charges()
	if len(charges) != 1 || charges[0].Amount != 33363 {
		t.Errorf("charges = %+v, want one of 33363", charges)
	}
}

func TestCheckoutRequiresAuthenticationAndNonEmptyCart(t *testing.T) {
	h := testutil.NewHarness(t)
	buyer := h.CreateUser(t, "Audience")

	requireError(t, h.GraphQLRaw(t, "", checkoutMutation, checkoutVars()), "authentication required")
	requireError(t, h.GraphQLRaw(t, "", myOrdersQuery, nil), "authentication required")
	requireError(t, h.GraphQLRaw(t, buyer.Token, checkoutMutation, checkoutVars()), "cart is empty")
	if n := len(h.Payments.Charges()); n != 0 {
		t.Errorf("charges = %d, want 0", n)
	}
}

func TestCheckoutRequiresCompleteShippingAddress(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 1)

	blank := address()
	blank["city"] = "   "
	requireError(t, h.GraphQLRaw(t, buyer.Token, checkoutMutation, map[string]any{"a": blank}), "shipping address is incomplete")

	missing := address()
	delete(missing, "country")
	if res := h.GraphQLRaw(t, buyer.Token, checkoutMutation, map[string]any{"a": missing}); len(res.Errors) == 0 {
		t.Error("missing country: expected an error")
	}
	if stockOf(t, h, vase) != 5 || len(myCart(t, h, buyer).Lines) != 1 {
		t.Error("rejected checkout changed stock or cart")
	}
}

func TestCheckoutFailsWholeWhenAnyLineIsUnavailable(t *testing.T) {
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

			requireError(t, h.GraphQLRaw(t, buyer.Token, checkoutMutation, checkoutVars()), "the-offender")

			if got := stockOf(t, h, fine); got != 5 {
				t.Errorf("fine stock = %d, want 5 (untouched)", got)
			}
			if got := stockOf(t, h, bad); got != badStock {
				t.Errorf("offender stock = %d, want %d", got, badStock)
			}
			if got := len(myCart(t, h, buyer).Lines); got != 2 {
				t.Errorf("cart lines = %d, want 2", got)
			}
			var orders struct {
				MyOrders []orderData `json:"myOrders"`
			}
			h.GraphQL(t, buyer.Token, myOrdersQuery, nil, &orders)
			if len(orders.MyOrders) != 0 {
				t.Errorf("orders = %+v, want none", orders.MyOrders)
			}
			if n := len(h.Payments.Charges()); n != 0 {
				t.Errorf("charges = %d, want 0", n)
			}
		})
	}
}

func TestPaymentFailureRollsEverythingBack(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)
	h.Payments.FailCharges(errors.New("card declined"))

	requireError(t, h.GraphQLRaw(t, buyer.Token, checkoutMutation, checkoutVars()), "payment failed")

	if got := stockOf(t, h, vase); got != 5 {
		t.Errorf("stock = %d, want 5", got)
	}
	if cart := myCart(t, h, buyer); len(cart.Lines) != 1 || cart.Lines[0].Quantity != 2 {
		t.Errorf("cart = %+v, want intact", cart.Lines)
	}
	var n int
	if err := h.DB.QueryRow(`SELECT (SELECT count(*) FROM orders) + (SELECT count(*) FROM order_items)`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("order rows = %d, want 0", n)
	}

	// The buyer can simply retry once payments work again.
	h.Payments.FailCharges(nil)
	var data struct {
		Checkout orderData `json:"checkout"`
	}
	h.GraphQL(t, buyer.Token, checkoutMutation, checkoutVars(), &data)
	if data.Checkout.State != "PAID" {
		t.Errorf("retry state = %s, want PAID", data.Checkout.State)
	}
}

func TestTwoBuyersRacingForLastUnitExactlyOneWins(t *testing.T) {
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
			results[i] = h.GraphQLRaw(t, users[i].Token, checkoutMutation, checkoutVars())
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
		t.Fatalf("successful checkouts = %d, want exactly 1", wins)
	}
	if got := stockOf(t, h, piece); got != 0 {
		t.Errorf("stock = %d, want 0", got)
	}
	if n := len(h.Payments.Charges()); n != 1 {
		t.Errorf("charges = %d, want 1", n)
	}
}

func TestOrderItemsAreSnapshotsUnaffectedByListingChanges(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "vase", "price": 50, "stock": 3})
	addToCart(t, h, buyer, vase, 1)
	var placed struct {
		Checkout orderData `json:"checkout"`
	}
	h.GraphQL(t, buyer.Token, checkoutMutation, checkoutVars(), &placed)

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
	if o.ID != placed.Checkout.ID || o.Total != 50 || o.State != "PAID" ||
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
		h.GraphQL(t, buyer.Token, checkoutMutation, checkoutVars(), nil)
	}
	addToCart(t, h, other, first, 1)
	h.GraphQL(t, other.Token, checkoutMutation, checkoutVars(), nil)

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
	h.GraphQL(t, buyer.Token, checkoutMutation, checkoutVars(), nil)

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
