package app_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kalasetu/testutil"
)

type cartLine struct {
	Listing struct {
		ID    string  `json:"id"`
		Title string  `json:"title"`
		Price float64 `json:"price"`
		Stock int     `json:"stock"`
	} `json:"listing"`
	Quantity int     `json:"quantity"`
	Issue    *string `json:"issue"`
}

type cartData struct {
	Lines []cartLine `json:"lines"`
	Total float64    `json:"total"`
}

const cartFields = `lines { quantity issue listing { id title price stock } } total`

const (
	myCartQuery        = `{ myCart { ` + cartFields + ` } }`
	addToCartMutation  = `mutation($id: ID!, $q: Int) { addToCart(listingId: $id, quantity: $q) { ` + cartFields + ` } }`
	updateCartMutation = `mutation($id: ID!, $q: Int!) { updateCartItem(listingId: $id, quantity: $q) { ` + cartFields + ` } }`
	removeCartMutation = `mutation($id: ID!) { removeFromCart(listingId: $id) { ` + cartFields + ` } }`
)

func myCart(t *testing.T, h *testutil.Harness, u testutil.User) cartData {
	t.Helper()
	var data struct {
		MyCart cartData `json:"myCart"`
	}
	h.GraphQL(t, u.Token, myCartQuery, nil, &data)
	return data.MyCart
}

func addToCart(t *testing.T, h *testutil.Harness, u testutil.User, listingID string, qty int) cartData {
	t.Helper()
	var data struct {
		AddToCart cartData `json:"addToCart"`
	}
	h.GraphQL(t, u.Token, addToCartMutation, map[string]any{"id": listingID, "q": qty}, &data)
	return data.AddToCart
}

func TestBuyerAddsListingToCartAndSeesItInMyCart(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"title": "vase", "price": 100.25, "stock": 5})

	returned := addToCart(t, h, buyer, vase, 2)
	got := myCart(t, h, buyer)

	for name, cart := range map[string]cartData{"addToCart": returned, "myCart": got} {
		if len(cart.Lines) != 1 {
			t.Fatalf("%s lines = %+v, want 1", name, cart.Lines)
		}
		line := cart.Lines[0]
		if line.Listing.ID != vase || line.Listing.Title != "vase" || line.Quantity != 2 || line.Issue != nil {
			t.Errorf("%s line = %+v, want 2 x vase with no issue", name, line)
		}
		if cart.Total != 200.5 {
			t.Errorf("%s total = %v, want 200.5", name, cart.Total)
		}
	}
}

func TestCartOperationsRequireAuthentication(t *testing.T) {
	h := testutil.NewHarness(t)
	id := map[string]any{"id": "1"}

	requireError(t, h.GraphQLRaw(t, "", myCartQuery, nil), "authentication required")
	requireError(t, h.GraphQLRaw(t, "", addToCartMutation, id), "authentication required")
	requireError(t, h.GraphQLRaw(t, "", updateCartMutation, map[string]any{"id": "1", "q": 1}), "authentication required")
	requireError(t, h.GraphQLRaw(t, "", removeCartMutation, id), "authentication required")
}

func TestAddingListingAlreadyInCartIncreasesQuantity(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})

	addToCart(t, h, buyer, vase, 2)
	got := addToCart(t, h, buyer, vase, 1)

	if len(got.Lines) != 1 || got.Lines[0].Quantity != 3 {
		t.Fatalf("lines = %+v, want a single line of quantity 3", got.Lines)
	}
}

func TestAddToCartDefaultsToQuantityOne(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})

	var data struct {
		AddToCart cartData `json:"addToCart"`
	}
	h.GraphQL(t, buyer.Token, `mutation($id: ID!) { addToCart(listingId: $id) { `+cartFields+` } }`, map[string]any{"id": vase}, &data)
	if len(data.AddToCart.Lines) != 1 || data.AddToCart.Lines[0].Quantity != 1 {
		t.Fatalf("lines = %+v, want quantity 1", data.AddToCart.Lines)
	}
}

func TestAddToCartRejectsInvalidAdditions(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	cat := categoryID(t, h, "Pottery")
	live := seedListing(t, h, seller, cat, map[string]any{"stock": 2})
	gone := seedListing(t, h, seller, cat, map[string]any{"stock": 2})
	archive(t, h, gone)
	soldOut := seedListing(t, h, seller, cat, map[string]any{"stock": 0})

	cases := []struct {
		name    string
		user    testutil.User
		id      string
		qty     int
		wantErr string
	}{
		{"own listing", seller, live, 1, "own listing"},
		{"archived listing", buyer, gone, 1, "archived"},
		{"sold out listing", buyer, soldOut, 1, "stock"},
		{"more than stock", buyer, live, 3, "stock"},
		{"zero quantity", buyer, live, 0, "quantity"},
		{"negative quantity", buyer, live, -1, "quantity"},
		{"unknown listing", buyer, "999999", 1, "not found"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireError(t, h.GraphQLRaw(t, c.user.Token, addToCartMutation, map[string]any{"id": c.id, "q": c.qty}), c.wantErr)
		})
	}
	if cart := myCart(t, h, buyer); len(cart.Lines) != 0 {
		t.Errorf("buyer cart = %+v, want empty after rejected additions", cart.Lines)
	}
}

func TestAddToCartCannotExceedStockAcrossSeveralAdds(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 3})

	addToCart(t, h, buyer, vase, 2)
	requireError(t, h.GraphQLRaw(t, buyer.Token, addToCartMutation, map[string]any{"id": vase, "q": 2}), "stock")
	if got := myCart(t, h, buyer); got.Lines[0].Quantity != 2 {
		t.Errorf("quantity = %d, want unchanged 2", got.Lines[0].Quantity)
	}
}

func updateCart(t *testing.T, h *testutil.Harness, u testutil.User, listingID string, qty int) cartData {
	t.Helper()
	var data struct {
		UpdateCartItem cartData `json:"updateCartItem"`
	}
	h.GraphQL(t, u.Token, updateCartMutation, map[string]any{"id": listingID, "q": qty}, &data)
	return data.UpdateCartItem
}

func TestUpdateCartItemSetsQuantity(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 4)

	got := updateCart(t, h, buyer, vase, 2)

	if len(got.Lines) != 1 || got.Lines[0].Quantity != 2 {
		t.Fatalf("lines = %+v, want quantity 2", got.Lines)
	}
	if again := myCart(t, h, buyer); again.Lines[0].Quantity != 2 {
		t.Errorf("myCart quantity = %d, want 2", again.Lines[0].Quantity)
	}
}

func TestUpdateCartItemWithZeroRemovesTheLine(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	cat := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, cat, map[string]any{"stock": 5})
	bowl := seedListing(t, h, seller, cat, map[string]any{"title": "bowl", "stock": 5})
	addToCart(t, h, buyer, vase, 1)
	addToCart(t, h, buyer, bowl, 1)

	got := updateCart(t, h, buyer, vase, 0)

	if len(got.Lines) != 1 || got.Lines[0].Listing.ID != bowl {
		t.Fatalf("lines = %+v, want only the bowl", got.Lines)
	}
}

func TestUpdateCartItemRejectsInvalidQuantities(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	cat := categoryID(t, h, "Pottery")
	vase := seedListing(t, h, seller, cat, map[string]any{"stock": 3})
	other := seedListing(t, h, seller, cat, map[string]any{"stock": 3})
	addToCart(t, h, buyer, vase, 2)

	cases := []struct {
		name    string
		id      string
		qty     int
		wantErr string
	}{
		{"negative", vase, -1, "quantity"},
		{"more than stock", vase, 4, "stock"},
		{"listing not in cart", other, 1, "not in your cart"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireError(t, h.GraphQLRaw(t, buyer.Token, updateCartMutation, map[string]any{"id": c.id, "q": c.qty}), c.wantErr)
		})
	}
	if got := myCart(t, h, buyer); len(got.Lines) != 1 || got.Lines[0].Quantity != 2 {
		t.Errorf("cart = %+v, want the original single line of 2", got.Lines)
	}
}

func TestRemoveFromCartDropsTheLineAndIsIdempotent(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)

	for i := 0; i < 2; i++ {
		var data struct {
			RemoveFromCart cartData `json:"removeFromCart"`
		}
		h.GraphQL(t, buyer.Token, removeCartMutation, map[string]any{"id": vase}, &data)
		if len(data.RemoveFromCart.Lines) != 0 || data.RemoveFromCart.Total != 0 {
			t.Fatalf("removal %d: cart = %+v, want empty", i+1, data.RemoveFromCart)
		}
	}
}

func TestMyCartTotalIsRoundedToTwoDecimals(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	cat := categoryID(t, h, "Pottery")
	// Summed as float64, 0.1*3 + 0.2 is 0.5000000000000001.
	a := seedListing(t, h, seller, cat, map[string]any{"price": 0.1, "stock": 5})
	b := seedListing(t, h, seller, cat, map[string]any{"price": 0.2, "stock": 5})
	addToCart(t, h, buyer, a, 3)
	addToCart(t, h, buyer, b, 1)

	if got := myCart(t, h, buyer); got.Total != 0.5 {
		t.Errorf("total = %v, want 0.5", got.Total)
	}
}

func TestEmptyCartHasNoLinesAndZeroTotal(t *testing.T) {
	h := testutil.NewHarness(t)
	buyer := h.CreateUser(t, "Audience")

	got := myCart(t, h, buyer)
	if got.Lines == nil || len(got.Lines) != 0 || got.Total != 0 {
		t.Errorf("cart = %#v, want empty non-nil lines and total 0", got)
	}
}

func TestMyCartFlagsLinesThatBecameUnavailableWithoutDroppingThem(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	cat := categoryID(t, h, "Pottery")
	fine := seedListing(t, h, seller, cat, map[string]any{"title": "fine", "price": 10, "stock": 5})
	archived := seedListing(t, h, seller, cat, map[string]any{"title": "archived", "price": 20, "stock": 5})
	soldOut := seedListing(t, h, seller, cat, map[string]any{"title": "soldout", "price": 30, "stock": 5})
	shrunk := seedListing(t, h, seller, cat, map[string]any{"title": "shrunk", "price": 40, "stock": 5})
	for _, id := range []string{fine, archived, soldOut, shrunk} {
		addToCart(t, h, buyer, id, 2)
	}

	archive(t, h, archived)
	if _, err := h.DB.Exec(`UPDATE listings SET stock = 0 WHERE id = $1`, soldOut); err != nil {
		t.Fatal(err)
	}
	if _, err := h.DB.Exec(`UPDATE listings SET stock = 1 WHERE id = $1`, shrunk); err != nil {
		t.Fatal(err)
	}

	got := myCart(t, h, buyer)

	want := map[string]string{"fine": "", "archived": "ARCHIVED", "soldout": "OUT_OF_STOCK", "shrunk": "INSUFFICIENT_STOCK"}
	if len(got.Lines) != len(want) {
		t.Fatalf("lines = %+v, want all %d lines kept", got.Lines, len(want))
	}
	for _, l := range got.Lines {
		issue := ""
		if l.Issue != nil {
			issue = *l.Issue
		}
		if issue != want[l.Listing.Title] {
			t.Errorf("%s issue = %q, want %q", l.Listing.Title, issue, want[l.Listing.Title])
		}
	}
	// Only the purchasable line counts: 2 x 10.
	if got.Total != 20 {
		t.Errorf("total = %v, want 20", got.Total)
	}
}

func TestFlaggedLineCanStillBeRemovedOrReduced(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	cat := categoryID(t, h, "Pottery")
	gone := seedListing(t, h, seller, cat, map[string]any{"stock": 5})
	shrunk := seedListing(t, h, seller, cat, map[string]any{"stock": 5})
	addToCart(t, h, buyer, gone, 2)
	addToCart(t, h, buyer, shrunk, 3)
	archive(t, h, gone)
	if _, err := h.DB.Exec(`UPDATE listings SET stock = 1 WHERE id = $1`, shrunk); err != nil {
		t.Fatal(err)
	}

	updateCart(t, h, buyer, shrunk, 1)
	got := updateCart(t, h, buyer, gone, 0)

	if len(got.Lines) != 1 || got.Lines[0].Listing.ID != shrunk || got.Lines[0].Issue != nil {
		t.Fatalf("lines = %+v, want only the reduced, now-valid line", got.Lines)
	}
}

func TestCartIsPrivateAndPersistsAcrossSessions(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Audience")
	stranger := h.CreateUser(t, "Audience")
	vase := seedListing(t, h, seller, categoryID(t, h, "Pottery"), map[string]any{"stock": 5})
	addToCart(t, h, buyer, vase, 2)

	if got := myCart(t, h, stranger); len(got.Lines) != 0 {
		t.Errorf("stranger sees %+v, want their own empty cart", got.Lines)
	}
	// A stranger's edits do not touch the buyer's cart, even by listing id.
	requireError(t, h.GraphQLRaw(t, stranger.Token, updateCartMutation, map[string]any{"id": vase, "q": 5}), "not in your cart")
	h.GraphQLRaw(t, stranger.Token, removeCartMutation, map[string]any{"id": vase})

	// A fresh login is a new session with a new token.
	relogin := loginAs(t, h, buyer)
	got := myCart(t, h, relogin)
	if len(got.Lines) != 1 || got.Lines[0].Quantity != 2 {
		t.Errorf("after re-login cart = %+v, want the original line of 2", got.Lines)
	}
}

// loginAs signs the user in again through the real login endpoint, returning
// the same user with a token from that new session.
func loginAs(t *testing.T, h *testutil.Harness, u testutil.User) testutil.User {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": u.Email, "password": "password123"})
	rec := httptest.NewRecorder()
	h.App.Router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("login: status %d: %s", rec.Code, rec.Body)
	}
	var res struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil || res.AccessToken == "" {
		t.Fatalf("decode login response: %v: %s", err, rec.Body)
	}
	u.Token = res.AccessToken
	return u
}
