package app_test

import (
	"fmt"
	"strings"
	"testing"

	"kalasetu/testutil"
)

func TestCategoriesArePublicAndSeeded(t *testing.T) {
	h := testutil.NewHarness(t)

	var data struct {
		Categories []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"categories"`
	}
	h.GraphQL(t, "", `{ categories { id name } }`, nil, &data)

	if len(data.Categories) < 3 {
		t.Fatalf("expected several seeded categories, got %+v", data.Categories)
	}
	var found bool
	for _, c := range data.Categories {
		if c.Name == "Pottery" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Pottery among categories, got %+v", data.Categories)
	}
}

const createListingMutation = `
mutation($input: CreateListingInput!) {
  createListing(input: $input) { id }
}`

func categoryID(t *testing.T, h *testutil.Harness, name string) string {
	t.Helper()
	var data struct {
		Categories []struct{ ID, Name string } `json:"categories"`
	}
	h.GraphQL(t, "", `{ categories { id name } }`, nil, &data)
	for _, c := range data.Categories {
		if c.Name == name {
			return c.ID
		}
	}
	t.Fatalf("category %q not found", name)
	return ""
}

func listingInput(catID string, overrides map[string]any) map[string]any {
	in := map[string]any{
		"title":       "Blue clay vase",
		"description": "Hand thrown and glazed",
		"price":       499.5,
		"stock":       1,
		"imageUrls":   []string{"https://img.test/cover.jpg", "https://img.test/side.jpg"},
		"categoryId":  catID,
	}
	for k, v := range overrides {
		in[k] = v
	}
	return map[string]any{"input": in}
}

func onboard(t *testing.T, h *testutil.Harness, u testutil.User, role string) {
	t.Helper()
	var out struct {
		OnboardUser bool `json:"onboardUser"`
	}
	h.GraphQL(t, u.Token, `mutation($i: OnboardingInput!) { onboardUser(input: $i) }`, map[string]any{
		"i": map[string]any{
			"name": "Asha Potter", "role": role, "location": "Jaipur",
			"profilePicture": "https://img.test/asha.jpg",
		},
	}, &out)
}

func TestSellerCreatesListingAndAnyoneViewsIt(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	onboard(t, h, seller, "Artist")
	cat := categoryID(t, h, "Pottery")

	var created struct {
		CreateListing struct{ ID string } `json:"createListing"`
	}
	h.GraphQL(t, seller.Token, createListingMutation, listingInput(cat, nil), &created)
	if created.CreateListing.ID == "" {
		t.Fatal("expected a listing id")
	}

	var got struct {
		Listing struct {
			ID          string
			Title       string
			Description string
			Price       float64
			Currency    string
			Stock       int
			ImageUrls   []string
			Category    struct{ Name string }
			Seller      struct {
				ID             string
				Name           string
				Location       string
				ProfilePicture string
			}
		} `json:"listing"`
	}
	h.GraphQL(t, "", `query($id: ID!) { listing(id: $id) {
		id title description price currency stock imageUrls category { name }
		seller { id name location profilePicture }
	} }`, map[string]any{"id": created.CreateListing.ID}, &got)

	l := got.Listing
	if l.Title != "Blue clay vase" || l.Description != "Hand thrown and glazed" {
		t.Errorf("unexpected title/description: %+v", l)
	}
	if l.Price != 499.5 || l.Currency != "INR" || l.Stock != 1 {
		t.Errorf("unexpected price/currency/stock: %v %s %d", l.Price, l.Currency, l.Stock)
	}
	if len(l.ImageUrls) != 2 || l.ImageUrls[0] != "https://img.test/cover.jpg" {
		t.Errorf("images = %v, want cover first", l.ImageUrls)
	}
	if l.Category.Name != "Pottery" {
		t.Errorf("category = %q", l.Category.Name)
	}
	if l.Seller.Name != "Asha Potter" || l.Seller.Location != "Jaipur" ||
		l.Seller.ProfilePicture != "https://img.test/asha.jpg" {
		t.Errorf("unexpected seller: %+v", l.Seller)
	}
}

func TestCreateListingRequiresAuthentication(t *testing.T) {
	h := testutil.NewHarness(t)
	cat := categoryID(t, h, "Pottery")

	res := h.GraphQLRaw(t, "", createListingMutation, listingInput(cat, nil))
	if len(res.Errors) == 0 {
		t.Fatal("expected an error for an unauthenticated createListing")
	}
}

func TestCreateListingRejectsUsersWithoutSellerRole(t *testing.T) {
	h := testutil.NewHarness(t)
	cat := categoryID(t, h, "Pottery")

	for _, role := range []string{"Organizer", "Sponsor"} {
		u := h.CreateUser(t, role)
		res := h.GraphQLRaw(t, u.Token, createListingMutation, listingInput(cat, nil))
		if len(res.Errors) == 0 {
			t.Errorf("role %s: expected createListing to be rejected", role)
		}
	}
}

func TestCraftspersonCanCreateListing(t *testing.T) {
	h := testutil.NewHarness(t)
	u := h.CreateUser(t, "Craftsperson")
	cat := categoryID(t, h, "Woodwork")

	var out struct {
		CreateListing struct{ ID string } `json:"createListing"`
	}
	h.GraphQL(t, u.Token, createListingMutation, listingInput(cat, nil), &out)
	if out.CreateListing.ID == "" {
		t.Fatal("expected a listing id")
	}
}

func imageURLs(n int) []string {
	urls := make([]string, n)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://img.test/%d.jpg", i)
	}
	return urls
}

func TestCreateListingValidation(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")

	invalid := map[string]map[string]any{
		"empty title":          {"title": ""},
		"blank title":          {"title": "   "},
		"zero price":           {"price": 0},
		"negative price":       {"price": -5},
		"price rounds to 0":    {"price": 0.004},
		"price too large":      {"price": 1e12},
		"negative stock":       {"stock": -1},
		"no images":            {"imageUrls": []string{}},
		"too many images":      {"imageUrls": imageURLs(9)},
		"blank image url":      {"imageUrls": []string{" "}},
		"unknown category":     {"categoryId": "999999"},
		"non-numeric category": {"categoryId": "abc"},
	}
	for name, override := range invalid {
		t.Run("rejects "+name, func(t *testing.T) {
			res := h.GraphQLRaw(t, seller.Token, createListingMutation, listingInput(cat, override))
			if len(res.Errors) == 0 {
				t.Fatalf("expected an error, got data %s", res.Data)
			}
		})
	}

	valid := map[string]map[string]any{
		"zero stock (sold out)": {"stock": 0},
		"eight images":          {"imageUrls": imageURLs(8)},
		"single image":          {"imageUrls": imageURLs(1)},
		"smallest price":        {"price": 0.01},
		"no description":        {"description": nil},
	}
	for name, override := range valid {
		t.Run("accepts "+name, func(t *testing.T) {
			res := h.GraphQLRaw(t, seller.Token, createListingMutation, listingInput(cat, override))
			if len(res.Errors) > 0 {
				t.Fatalf("unexpected errors: %+v", res.Errors)
			}
		})
	}
}

func TestListingForUnknownIDIsNull(t *testing.T) {
	h := testutil.NewHarness(t)

	var data struct {
		Listing *struct{ ID string } `json:"listing"`
	}
	h.GraphQL(t, "", `{ listing(id: "424242") { id } }`, nil, &data)
	if data.Listing != nil {
		t.Fatalf("expected null, got %+v", data.Listing)
	}
}

func TestListingNeverExposesSellerEmail(t *testing.T) {
	h := testutil.NewHarness(t)
	seller := h.CreateUser(t, "Artist")
	cat := categoryID(t, h, "Pottery")
	var created struct {
		CreateListing struct{ ID string } `json:"createListing"`
	}
	h.GraphQL(t, seller.Token, createListingMutation, listingInput(cat, nil), &created)

	res := h.GraphQLRaw(t, "", `query($id: ID!) { listing(id: $id) { seller { email } } }`,
		map[string]any{"id": created.CreateListing.ID})
	if len(res.Errors) == 0 {
		t.Fatalf("expected seller.email to be rejected by the schema, got %s", res.Data)
	}

	// And the email appears nowhere in a full listing response.
	full := h.GraphQLRaw(t, "", `query($id: ID!) { listing(id: $id) {
		id title description price currency stock imageUrls category { id name }
		seller { id name location profilePicture } createdAt } }`,
		map[string]any{"id": created.CreateListing.ID})
	if strings.Contains(string(full.Data), seller.Email) {
		t.Fatalf("seller email leaked: %s", full.Data)
	}
}
