package models

import "time"

type CheckoutSessionStatus string

const (
	CheckoutSessionOpen      CheckoutSessionStatus = "OPEN"
	CheckoutSessionConsumed  CheckoutSessionStatus = "CONSUMED"
	CheckoutSessionCancelled CheckoutSessionStatus = "CANCELLED"
	CheckoutSessionExpired   CheckoutSessionStatus = "EXPIRED"
)

// CheckoutSessionItem holds a snapshot of the Listing's title, price and
// quantity at the moment the Checkout Session reserved it.
type CheckoutSessionItem struct {
	ID        int
	ListingID int
	SellerID  int
	Title     string
	Price     float64
	Quantity  int
}

// CheckoutSession is a Buyer's reserved, time-limited intent to purchase: a
// snapshot of the Cart lines and Shipping address, the total, and the
// gateway order id the Buyer's browser needs to open the payment screen.
type CheckoutSession struct {
	ID       int
	BuyerID  int
	Items    []CheckoutSessionItem
	Shipping ShippingAddress
	Total    float64
	// GatewayOrderID is the gateway's identifier for the matching payment order.
	GatewayOrderID string
	// KeyID is the gateway's public key, safe to hand to the Buyer's browser.
	// It is not persisted; it is filled in by the service from configuration.
	KeyID     string
	Status    CheckoutSessionStatus
	ExpiresAt time.Time
	CreatedAt time.Time
}
