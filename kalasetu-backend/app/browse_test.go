package app_test

import (
	"fmt"
	"reflect"
	"testing"

	"kalasetu/testutil"
)

type listingSummary struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// seedListing creates a listing through the GraphQL seam and returns its id.
func seedListing(t *testing.T, h *testutil.Harness, seller testutil.User, catID string, overrides map[string]any) string {
	t.Helper()
	var out struct {
		CreateListing struct{ ID string } `json:"createListing"`
	}
	h.GraphQL(t, seller.Token, createListingMutation, listingInput(catID, overrides), &out)
	return out.CreateListing.ID
}

func archive(t *testing.T, h *testutil.Harness, id string) {
	t.Helper()
	if _, err := h.DB.Exec(`UPDATE listings SET archived_at = now() WHERE id = $1`, id); err != nil {
		t.Fatalf("archive: %v", err)
	}
}

func browse(t *testing.T, h *testutil.Harness, args string) []listingSummary {
	t.Helper()
	return browseVars(t, h, args, nil)
}

func browseVars(t *testing.T, h *testutil.Harness, args string, vars map[string]any) []listingSummary {
	t.Helper()
	var data struct {
		Listings []listingSummary `json:"listings"`
	}
	decl := ""
	if vars != nil {
		decl = "($q: String)"
	}
	h.GraphQL(t, "", fmt.Sprintf(`query%s { listings%s { id title price stock } }`, decl, args), vars, &data)
	return data.Listings
}

func titles(ls []listingSummary) []string {
	out := make([]string, 0, len(ls))
	for _, l := range ls {
		out = append(out, l.Title)
	}
	return out
}

func assertTitles(t *testing.T, got []listingSummary, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if !reflect.DeepEqual(titles(got), want) {
		t.Fatalf("titles = %v, want %v", titles(got), want)
	}
}

func TestListingsAreBrowsablePubliclyNewestFirst(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	seedListing(t, h, seller, cat, map[string]any{"title": "first"})
	seedListing(t, h, seller, cat, map[string]any{"title": "second"})
	seedListing(t, h, seller, cat, map[string]any{"title": "third"})

	assertTitles(t, browse(t, h, ""), "third", "second", "first")
}

func TestListingsEmptyResultIsEmptyList(t *testing.T) {
	h := testutil.NewHarness(t)

	res := h.GraphQLRaw(t, "", `{ listings { id } }`, nil)
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %+v", res.Errors)
	}
	if string(res.Data) != `{"listings":[]}` {
		t.Fatalf("data = %s, want empty list", res.Data)
	}
}

func TestListingsKeywordSearchIsCaseInsensitiveOverTitleAndDescription(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	seedListing(t, h, seller, cat, map[string]any{"title": "Blue Vase", "description": "glazed"})
	seedListing(t, h, seller, cat, map[string]any{"title": "Plate", "description": "Painted BLUE by hand"})
	seedListing(t, h, seller, cat, map[string]any{"title": "Red bowl", "description": "plain"})

	assertTitles(t, browse(t, h, `(filter: {query: "bLuE"})`), "Plate", "Blue Vase")
	assertTitles(t, browse(t, h, `(filter: {query: "nomatch"})`))
}

func TestListingsSearchTreatsSpecialCharactersLiterally(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	seedListing(t, h, seller, cat, map[string]any{"title": "100% cotton", "description": "x"})
	seedListing(t, h, seller, cat, map[string]any{"title": "snake_case mug", "description": "x"})
	seedListing(t, h, seller, cat, map[string]any{"title": "snakeXcase jar", "description": "x"})
	seedListing(t, h, seller, cat, map[string]any{"title": "O'Brien's pot", "description": "x"})
	seedListing(t, h, seller, cat, map[string]any{"title": `back\slash`, "description": "x"})

	cases := map[string][]string{
		"%":                          {"100% cotton"},
		"_":                          {"snake_case mug"},
		"snake_case":                 {"snake_case mug"},
		"'":                          {"O'Brien's pot"},
		"O'Brien":                    {"O'Brien's pot"},
		`\`:                          {`back\slash`},
		"'; DROP TABLE listings; --": nil,
		`"`:                          nil,
	}
	for q, want := range cases {
		t.Run(q, func(t *testing.T) {
			got := browseVars(t, h, `(filter: {query: $q})`, map[string]any{"q": q})
			assertTitles(t, got, want...)
		})
	}
	// The table survived the injection attempt.
	if n := len(browse(t, h, "")); n != 5 {
		t.Fatalf("expected 5 listings to remain, got %d", n)
	}
}

func TestListingsFilterByCategory(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	pottery, wood := categoryID(t, h, "Pottery"), categoryID(t, h, "Woodwork")
	seedListing(t, h, seller, pottery, map[string]any{"title": "vase"})
	seedListing(t, h, seller, wood, map[string]any{"title": "spoon"})

	assertTitles(t, browse(t, h, fmt.Sprintf(`(filter: {categoryId: %q})`, wood)), "spoon")
}

func TestListingsFilterByPriceRangeInclusive(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	for _, p := range []float64{100, 200, 300, 400} {
		seedListing(t, h, seller, cat, map[string]any{"title": fmt.Sprintf("p%d", int(p)), "price": p})
	}

	assertTitles(t, browse(t, h, `(filter: {minPrice: 200, maxPrice: 300}, sort: PRICE_ASC)`), "p200", "p300")
	assertTitles(t, browse(t, h, `(filter: {minPrice: 300}, sort: PRICE_ASC)`), "p300", "p400")
	assertTitles(t, browse(t, h, `(filter: {maxPrice: 100})`), "p100")
	assertTitles(t, browse(t, h, `(filter: {minPrice: 500})`))
}

func TestListingsCombineFilters(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	pottery, wood := categoryID(t, h, "Pottery"), categoryID(t, h, "Woodwork")
	seedListing(t, h, seller, pottery, map[string]any{"title": "blue vase", "price": 100})
	seedListing(t, h, seller, pottery, map[string]any{"title": "blue jug", "price": 900})
	seedListing(t, h, seller, wood, map[string]any{"title": "blue spoon", "price": 100})

	got := browse(t, h, fmt.Sprintf(`(filter: {query: "blue", categoryId: %q, maxPrice: 500})`, pottery))
	assertTitles(t, got, "blue vase")
}

func TestListingsSorting(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	seedListing(t, h, seller, cat, map[string]any{"title": "mid", "price": 200})
	seedListing(t, h, seller, cat, map[string]any{"title": "cheap", "price": 100})
	seedListing(t, h, seller, cat, map[string]any{"title": "dear", "price": 300})

	assertTitles(t, browse(t, h, `(sort: NEWEST)`), "dear", "cheap", "mid")
	assertTitles(t, browse(t, h, `(sort: PRICE_ASC)`), "cheap", "mid", "dear")
	assertTitles(t, browse(t, h, `(sort: PRICE_DESC)`), "dear", "mid", "cheap")
}

func TestListingsPagination(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	for i := 1; i <= 5; i++ {
		seedListing(t, h, seller, cat, map[string]any{"title": fmt.Sprintf("item%d", i)})
	}

	assertTitles(t, browse(t, h, `(limit: 2)`), "item5", "item4")
	assertTitles(t, browse(t, h, `(limit: 2, offset: 2)`), "item3", "item2")
	assertTitles(t, browse(t, h, `(limit: 2, offset: 4)`), "item1")
	assertTitles(t, browse(t, h, `(limit: 2, offset: 50)`))
}

func TestListingsPaginationBounds(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")

	// Default page size is 20 and the cap is 100.
	for i := 0; i < 105; i++ {
		if _, err := h.DB.Exec(`
			INSERT INTO listings (seller_id, category_id, title, price, stock)
			VALUES ($1, $2, $3, 10, 1)`, seller.ID, cat, fmt.Sprintf("bulk%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(browse(t, h, "")); n != 20 {
		t.Errorf("default page size = %d, want 20", n)
	}
	if n := len(browse(t, h, `(limit: 100)`)); n != 100 {
		t.Errorf("limit 100 returned %d", n)
	}
	if n := len(browse(t, h, `(limit: 5000)`)); n != 100 {
		t.Errorf("limit above the maximum returned %d, want it capped at 100", n)
	}

	for _, args := range []string{`(limit: 0)`, `(limit: -1)`, `(offset: -1)`} {
		res := h.GraphQLRaw(t, "", `{ listings`+args+` { id } }`, nil)
		if len(res.Errors) == 0 {
			t.Errorf("%s: expected an error, got %s", args, res.Data)
		}
	}
}

func TestListingsNeverReturnArchived(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	seedListing(t, h, seller, cat, map[string]any{"title": "visible"})
	gone := seedListing(t, h, seller, cat, map[string]any{"title": "gone vase"})
	archive(t, h, gone)

	assertTitles(t, browse(t, h, ""), "visible")
	assertTitles(t, browse(t, h, `(filter: {query: "gone"})`))
	assertTitles(t, browse(t, h, fmt.Sprintf(`(filter: {categoryId: %q, minPrice: 1, maxPrice: 1000})`, cat)), "visible")
}

func TestListingsWithNoStockAreShownUnlessInStockOnly(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	seedListing(t, h, seller, cat, map[string]any{"title": "in stock", "stock": 3})
	seedListing(t, h, seller, cat, map[string]any{"title": "sold out", "stock": 0})

	got := browse(t, h, "")
	assertTitles(t, got, "sold out", "in stock")
	if got[0].Stock != 0 {
		t.Errorf("sold-out listing should report stock 0, got %d", got[0].Stock)
	}

	assertTitles(t, browse(t, h, `(filter: {inStockOnly: true})`), "in stock")
	assertTitles(t, browse(t, h, `(filter: {inStockOnly: false})`), "sold out", "in stock")
}

func featured(t *testing.T, h *testutil.Harness, args string) []listingSummary {
	t.Helper()
	var data struct {
		FeaturedListings []listingSummary `json:"featuredListings"`
	}
	h.GraphQL(t, "", `{ featuredListings`+args+` { id title price stock } }`, nil, &data)
	return data.FeaturedListings
}

func TestFeaturedListingsRespectLimitAndExcludeArchivedAndSoldOut(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	for i := 0; i < 6; i++ {
		seedListing(t, h, seller, cat, map[string]any{"title": fmt.Sprintf("live%d", i)})
	}
	archive(t, h, seedListing(t, h, seller, cat, map[string]any{"title": "archived"}))
	seedListing(t, h, seller, cat, map[string]any{"title": "sold out", "stock": 0})

	if n := len(featured(t, h, `(limit: 3)`)); n != 3 {
		t.Errorf("limit 3 returned %d", n)
	}
	all := featured(t, h, `(limit: 50)`)
	if len(all) != 6 {
		t.Fatalf("expected the 6 live listings, got %v", titles(all))
	}
	for _, l := range all {
		if l.Title == "archived" || l.Title == "sold out" {
			t.Errorf("featured must not include %q", l.Title)
		}
	}
}

func TestFeaturedListingsAreInRandomOrder(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	for i := 0; i < 8; i++ {
		seedListing(t, h, seller, cat, map[string]any{"title": fmt.Sprintf("live%d", i)})
	}

	first := titles(featured(t, h, `(limit: 8)`))
	for i := 0; i < 30; i++ {
		if !reflect.DeepEqual(first, titles(featured(t, h, `(limit: 8)`))) {
			return
		}
	}
	t.Fatalf("30 calls returned the identical order %v; expected it to vary", first)
}

func TestFeaturedListingsEmptyAndBounds(t *testing.T) {
	h := testutil.NewHarness(t)

	res := h.GraphQLRaw(t, "", `{ featuredListings { id } }`, nil)
	if len(res.Errors) > 0 || string(res.Data) != `{"featuredListings":[]}` {
		t.Fatalf("empty: data=%s errors=%+v", res.Data, res.Errors)
	}
	for _, args := range []string{`(limit: 0)`, `(limit: -3)`} {
		if res := h.GraphQLRaw(t, "", `{ featuredListings`+args+` { id } }`, nil); len(res.Errors) == 0 {
			t.Errorf("%s: expected an error", args)
		}
	}

	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	for i := 0; i < 3; i++ {
		seedListing(t, h, seller, cat, nil)
	}
	if n := len(featured(t, h, `(limit: 100000)`)); n != 3 {
		t.Errorf("huge limit returned %d, want 3 (capped, not an error)", n)
	}
}
