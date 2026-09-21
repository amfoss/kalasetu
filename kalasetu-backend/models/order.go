package models

import (
	"fmt"
	"time"
)

type FulfilmentStatus string

const (
	StatusPaid      FulfilmentStatus = "PAID"
	StatusShipped   FulfilmentStatus = "SHIPPED"
	StatusDelivered FulfilmentStatus = "DELIVERED"
	StatusCancelled FulfilmentStatus = "CANCELLED"
)

// CanAdvanceTo reports whether a Seller may move an Order Item from s to next:
// paid to shipped to delivered, one step at a time and never backwards.
func (s FulfilmentStatus) CanAdvanceTo(next FulfilmentStatus) bool {
	return (s == StatusPaid && next == StatusShipped) || (s == StatusShipped && next == StatusDelivered)
}

// OrderState is derived from an Order's items; the Order stores no status.
type OrderState string

const (
	OrderPaid             OrderState = "PAID"
	OrderPartiallyShipped OrderState = "PARTIALLY_SHIPPED"
	OrderShipped          OrderState = "SHIPPED"
	OrderDelivered        OrderState = "DELIVERED"
	OrderCancelled        OrderState = "CANCELLED"
)

type ShippingAddress struct {
	Name       string
	Phone      string
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    string
}

// OrderItem holds a snapshot of the Listing's title and price (INR).
type OrderItem struct {
	ID        int
	ListingID int
	SellerID  int
	Title     string
	Price     float64
	Quantity  int
	Status    FulfilmentStatus
}

// Order's Total is INR, rounded to 2 decimals.
type Order struct {
	ID        int
	BuyerID   int
	Items     []OrderItem
	Shipping  ShippingAddress
	Total     float64
	CreatedAt time.Time
}

// SellerOrderItem is an OrderItem seen from its Seller's side: it carries the
// Order's Shipping address.
type SellerOrderItem struct {
	OrderItem
	OrderID   int
	BuyerID   int
	Shipping  ShippingAddress
	CreatedAt time.Time
}

// State summarises the items that are not cancelled.
func (o Order) State() OrderState {
	var live, shipped, delivered int
	for _, it := range o.Items {
		if it.Status == StatusCancelled {
			continue
		}
		switch it.Status {
		case StatusShipped:
			shipped++
		case StatusDelivered:
			delivered++
		}
		live++
	}
	switch {
	case live == 0:
		return OrderCancelled
	case delivered == live:
		return OrderDelivered
	case shipped+delivered == live:
		return OrderShipped
	case shipped+delivered > 0:
		return OrderPartiallyShipped
	}
	// Nothing is shipped or delivered, so every live item is paid.
	return OrderPaid
}

// ListingUnavailableError is returned by checkout when a Cart line cannot be
// bought: the Listing is out of stock, short of stock, or Archived.
type ListingUnavailableError struct {
	ListingID int
	Title     string
}

func (e *ListingUnavailableError) Error() string {
	return fmt.Sprintf("%q (listing %d) is archived or does not have enough stock", e.Title, e.ListingID)
}
