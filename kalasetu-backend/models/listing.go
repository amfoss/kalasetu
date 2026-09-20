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
