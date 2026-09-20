package app_test

import (
	"strings"
	"testing"

	"kalasetu/testutil"
)

type myListing struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Stock     int      `json:"stock"`
	Archived  bool     `json:"archived"`
	Price     float64  `json:"price"`
	ImageUrls []string `json:"imageUrls"`
	Category  struct{ Name string }
}

const updateListingMutation = `
mutation($id: ID!, $input: UpdateListingInput!) {
  updateListing(id: $id, input: $input) { id title description price stock imageUrls category { name } }
}`

const archiveListingMutation = `mutation($id: ID!) { archiveListing(id: $id) }`

func myListings(t *testing.T, h *testutil.Harness, u testutil.User) []myListing {
	t.Helper()
	var data struct {
		MyListings []myListing `json:"myListings"`
	}
	h.GraphQL(t, u.Token, `{ myListings { id title stock archived price imageUrls category { name } } }`, nil, &data)
	return data.MyListings
}

func archiveViaAPI(t *testing.T, h *testutil.Harness, u testutil.User, id string) {
	t.Helper()
	var out struct {
		ArchiveListing bool `json:"archiveListing"`
	}
	h.GraphQL(t, u.Token, archiveListingMutation, map[string]any{"id": id}, &out)
	if !out.ArchiveListing {
		t.Fatal("archiveListing returned false")
	}
}

func requireError(t *testing.T, res testutil.GraphQLResponse, contains string) {
	t.Helper()
	if len(res.Errors) == 0 {
		t.Fatalf("expected an error containing %q, got data %s", contains, res.Data)
	}
	if !strings.Contains(res.Errors[0].Message, contains) {
		t.Fatalf("error = %q, want it to contain %q", res.Errors[0].Message, contains)
	}
}

func TestMyListingsReturnsOnlyCallersListingsIncludingArchived(t *testing.T) {
	h := testutil.NewHarness(t)
	mine := h.CreateUser(t, "Artist")
	other := h.CreateUser(t, "Craftsperson")
	cat := categoryID(t, h, "Pottery")

	seedListing(t, h, mine, cat, map[string]any{"title": "live", "stock": 3})
	gone := seedListing(t, h, mine, cat, map[string]any{"title": "gone", "stock": 0})
	seedListing(t, h, other, cat, map[string]any{"title": "theirs"})
	archiveViaAPI(t, h, mine, gone)

	got := myListings(t, h, mine)
	if len(got) != 2 {
		t.Fatalf("myListings = %+v, want 2", got)
	}
	// Newest first, so the archived one comes first.
	if got[0].Title != "gone" || !got[0].Archived || got[0].Stock != 0 {
		t.Errorf("first = %+v, want archived 'gone' with stock 0", got[0])
	}
	if got[1].Title != "live" || got[1].Archived || got[1].Stock != 3 {
		t.Errorf("second = %+v, want live 'live' with stock 3", got[1])
	}
}

func TestMyListingsRequiresAuthenticationAndIsEmptyForNewSeller(t *testing.T) {
	h := testutil.NewHarness(t)
	requireError(t, h.GraphQLRaw(t, "", `{ myListings { id } }`, nil), "authentication required")

	seller := h.CreateUser(t, "Artist")
	if got := myListings(t, h, seller); got == nil || len(got) != 0 {
		t.Fatalf("myListings = %#v, want empty non-nil list", got)
	}
}

func TestUpdateListingIsPartial(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	id := seedListing(t, h, seller, cat, nil)

	var out struct {
		UpdateListing struct {
			Title       string
			Description string
			Price       float64
			Stock       int
			ImageUrls   []string
			Category    struct{ Name string }
		} `json:"updateListing"`
	}
	h.GraphQL(t, seller.Token, updateListingMutation, map[string]any{
		"id": id, "input": map[string]any{"title": "Renamed vase", "stock": 7},
	}, &out)

	l := out.UpdateListing
	if l.Title != "Renamed vase" || l.Stock != 7 {
		t.Errorf("changed fields wrong: %+v", l)
	}
	if l.Description != "Hand thrown and glazed" || l.Price != 499.5 || l.Category.Name != "Pottery" ||
		len(l.ImageUrls) != 2 || l.ImageUrls[0] != "https://img.test/cover.jpg" {
		t.Errorf("untouched fields changed: %+v", l)
	}
}

func TestUpdateListingCanChangeEveryField(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	id := seedListing(t, h, seller, categoryID(t, h, "Pottery"), nil)
	wood := categoryID(t, h, "Woodwork")

	var out struct {
		UpdateListing struct {
			Title, Description string
			Price              float64
			Stock              int
			ImageUrls          []string
			Category           struct{ Name string }
		} `json:"updateListing"`
	}
	h.GraphQL(t, seller.Token, updateListingMutation, map[string]any{
		"id": id, "input": map[string]any{
			"title": "Oak bowl", "description": "", "price": 120.25, "stock": 0,
			"imageUrls": []string{"https://img.test/new.jpg"}, "categoryId": wood,
		},
	}, &out)

	l := out.UpdateListing
	if l.Title != "Oak bowl" || l.Description != "" || l.Price != 120.25 || l.Stock != 0 ||
		l.Category.Name != "Woodwork" || len(l.ImageUrls) != 1 || l.ImageUrls[0] != "https://img.test/new.jpg" {
		t.Errorf("update not applied: %+v", l)
	}
}

func TestUpdateListingAppliesCreationValidation(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	id := seedListing(t, h, seller, categoryID(t, h, "Pottery"), nil)

	invalid := map[string]map[string]any{
		"blank title":      {"title": "  "},
		"zero price":       {"price": 0},
		"price too large":  {"price": 1e12},
		"negative stock":   {"stock": -1},
		"no images":        {"imageUrls": []string{}},
		"too many images":  {"imageUrls": imageURLs(9)},
		"blank image url":  {"imageUrls": []string{" "}},
		"unknown category": {"categoryId": "999999"},
		"bad category id":  {"categoryId": "abc"},
	}
	for name, input := range invalid {
		t.Run("rejects "+name, func(t *testing.T) {
			res := h.GraphQLRaw(t, seller.Token, updateListingMutation, map[string]any{"id": id, "input": input})
			if len(res.Errors) == 0 {
				t.Fatalf("expected an error, got data %s", res.Data)
			}
		})
	}

	// A rejected update must leave the listing untouched.
	got := myListings(t, h, seller)
	if len(got) != 1 || got[0].Title != "Blue clay vase" || got[0].Price != 499.5 || got[0].Stock != 1 {
		t.Errorf("listing changed by rejected updates: %+v", got)
	}
}

func TestUpdateListingIsOwnerOnly(t *testing.T) {
	h := testutil.NewHarness(t)
	owner := h.CreateUser(t, "Artist")
	other := h.CreateUser(t, "Artist")
	id := seedListing(t, h, owner, categoryID(t, h, "Pottery"), nil)
	vars := map[string]any{"id": id, "input": map[string]any{"title": "hijacked"}}

	requireError(t, h.GraphQLRaw(t, other.Token, updateListingMutation, vars), "forbidden")
	requireError(t, h.GraphQLRaw(t, "", updateListingMutation, vars), "authentication required")

	if got := myListings(t, h, owner); got[0].Title != "Blue clay vase" {
		t.Errorf("title = %q, want unchanged", got[0].Title)
	}
}

func TestUpdateListingUnknownIDIsNotFound(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	requireError(t, h.GraphQLRaw(t, seller.Token, updateListingMutation,
		map[string]any{"id": "999999", "input": map[string]any{"title": "x"}}), "not found")
}

func TestArchiveListingIsOwnerOnly(t *testing.T) {
	h := testutil.NewHarness(t)
	owner := h.CreateUser(t, "Artist")
	other := h.CreateUser(t, "Artist")
	id := seedListing(t, h, owner, categoryID(t, h, "Pottery"), nil)
	vars := map[string]any{"id": id}

	requireError(t, h.GraphQLRaw(t, other.Token, archiveListingMutation, vars), "forbidden")
	requireError(t, h.GraphQLRaw(t, "", archiveListingMutation, vars), "authentication required")
	requireError(t, h.GraphQLRaw(t, owner.Token, archiveListingMutation, map[string]any{"id": "999999"}), "not found")

	if got := myListings(t, h, owner); got[0].Archived {
		t.Error("listing was archived by a rejected request")
	}
}

func TestArchivedListingVanishesFromPublicViews(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	buyer := h.CreateUser(t, "Sponsor")
	cat := categoryID(t, h, "Pottery")
	keep := seedListing(t, h, seller, cat, map[string]any{"title": "keep"})
	gone := seedListing(t, h, seller, cat, map[string]any{"title": "gone"})

	archiveViaAPI(t, h, seller, gone)

	assertTitles(t, browse(t, h, ""), "keep")
	if got := featured(t, h, "(limit: 100)"); len(got) != 1 || got[0].ID != keep {
		t.Errorf("featuredListings = %+v, want only 'keep'", got)
	}

	const q = `query($id: ID!) { listing(id: $id) { id } }`
	var data struct {
		Listing *struct{ ID string } `json:"listing"`
	}
	for name, token := range map[string]string{"anonymous": "", "other user": buyer.Token} {
		data.Listing = nil
		h.GraphQL(t, token, q, map[string]any{"id": gone}, &data)
		if data.Listing != nil {
			t.Errorf("%s can still see the archived listing", name)
		}
	}

	// The owner still sees it, both directly and in myListings.
	h.GraphQL(t, seller.Token, q, map[string]any{"id": gone}, &data)
	if data.Listing == nil {
		t.Error("owner should still be able to open their archived listing")
	}
	if mine := myListings(t, h, seller); len(mine) != 2 || !mine[0].Archived {
		t.Errorf("myListings = %+v, want both listings with the newest archived", mine)
	}
}

func TestArchivedListingCannotBeUpdated(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	id := seedListing(t, h, seller, categoryID(t, h, "Pottery"), nil)
	archiveViaAPI(t, h, seller, id)

	requireError(t, h.GraphQLRaw(t, seller.Token, updateListingMutation,
		map[string]any{"id": id, "input": map[string]any{"title": "revived"}}), "archived")
	if got := myListings(t, h, seller); got[0].Title != "Blue clay vase" {
		t.Errorf("archived listing was modified: %+v", got[0])
	}
}

func TestArchivingTwiceIsHarmless(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	id := seedListing(t, h, seller, categoryID(t, h, "Pottery"), nil)
	archiveViaAPI(t, h, seller, id)
	archiveViaAPI(t, h, seller, id)
	if got := myListings(t, h, seller); len(got) != 1 || !got[0].Archived {
		t.Errorf("myListings = %+v, want one archived listing", got)
	}
}
