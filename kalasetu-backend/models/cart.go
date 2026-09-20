package models

// CartItem is one line of a Buyer's Cart as stored.
type CartItem struct {
	ListingID int
	Quantity  int
}

type CartLineIssue string

const (
	CartIssueArchived          CartLineIssue = "ARCHIVED"
	CartIssueOutOfStock        CartLineIssue = "OUT_OF_STOCK"
	CartIssueInsufficientStock CartLineIssue = "INSUFFICIENT_STOCK"
)

// CartLine is a stored line joined with the Listing's current details. Issue is
// empty when the line can be bought as it stands.
type CartLine struct {
	Listing  Listing
	Quantity int
	Issue    CartLineIssue
}

// Cart's Total is INR, rounded to 2 decimals, and counts only lines with no issue.
type Cart struct {
	Lines []CartLine
	Total float64
}
