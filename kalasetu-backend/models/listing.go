package models

import "time"

type Category struct {
	ID   int
	Name string
}

// Listing is a seller's offering of a handmade item. Price is INR.
type Listing struct {
	ID          int
	Title       string
	Description string
	Price       float64
	Currency    string
	Stock       int
	ImageURLs   []string // first is the cover
	Category    Category
	Seller      ListingSeller
	CreatedAt   time.Time
	Archived    bool
}

// ListingSeller is the public part of a seller's profile; it deliberately has
// no email.
type ListingSeller struct {
	ID             int
	Name           string
	Location       string
	ProfilePicture string
}

type CreateListingInput struct {
	Title       string
	Description string
	Price       float64
	Stock       int
	ImageURLs   []string
	CategoryID  int
}

// UpdateListingInput is a partial update: nil fields are left unchanged.
type UpdateListingInput struct {
	Title       *string
	Description *string
	Price       *float64
	Stock       *int
	ImageURLs   []string // nil leaves images unchanged
	CategoryID  *int
}

type ListingSort string

const (
	SortNewest    ListingSort = "NEWEST"
	SortPriceAsc  ListingSort = "PRICE_ASC"
	SortPriceDesc ListingSort = "PRICE_DESC"
)

// ListingQuery describes a browse of non-archived Listings. Nil filter fields
// are unset.
type ListingQuery struct {
	Query       string
	CategoryID  *int
	MinPrice    *float64
	MaxPrice    *float64
	InStockOnly bool
	Sort        ListingSort
	Limit       int
	Offset      int
}
