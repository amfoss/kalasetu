package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"kalasetu/models"
)

// ErrLiveCheckoutSession means the Buyer already has an open Checkout
// Session; starting another is rejected rather than superseding it, because
// superseding could release Stock out from under a payment screen that is
// still open in the Buyer's browser.
var ErrLiveCheckoutSession = errors.New("you already have a checkout in progress")

// CreateOrderFunc registers a payment order with the gateway for amountCents
// and returns the gateway's order id. It runs inside the session-creation
// transaction, after Stock has been reserved.
type CreateOrderFunc func(ctx context.Context, amountCents int64, reference string) (gatewayOrderID string, err error)

type CheckoutSessionRepository interface {
	// Create reserves Stock for the Buyer's Cart and opens a Checkout Session,
	// all in a single transaction: the Cart is locked and validated, Stock is
	// decremented with the same guarded update Checkout uses, a session row
	// is inserted, createOrder registers the payment order with the gateway,
	// and the lines are snapshotted onto the session. Any failure, including
	// createOrder's, rolls everything back. It returns ErrLiveCheckoutSession
	// if the Buyer already has an open session, ErrCartEmpty, ErrCheckoutOwnBuy
	// or *models.ListingUnavailableError for business failures, and
	// createOrder's error unchanged.
	Create(ctx context.Context, buyerID int, ship models.ShippingAddress, expiresAt time.Time, createOrder CreateOrderFunc) (*models.CheckoutSession, error)
}

type checkoutSessionRepository struct {
	db *sql.DB
}

func NewCheckoutSessionRepository(db *sql.DB) CheckoutSessionRepository {
	return &checkoutSessionRepository{db: db}
}

func (r *checkoutSessionRepository) Create(ctx context.Context, buyerID int, ship models.ShippingAddress, expiresAt time.Time, createOrder CreateOrderFunc) (*models.CheckoutSession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var liveID int
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM checkout_sessions WHERE buyer_id = $1 AND status = 'open' FOR UPDATE`, buyerID).Scan(&liveID)
	if err == nil {
		return nil, ErrLiveCheckoutSession
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Locking the cart rows makes a double submit by the same buyer see an
	// empty cart; the fixed listing order keeps concurrent checkouts from
	// deadlocking on stock updates.
	rows, err := tx.QueryContext(ctx, `
		SELECT c.listing_id, c.quantity, l.title, (l.price * 100)::bigint, l.seller_id
		FROM cart_items c JOIN listings l ON l.id = c.listing_id
		WHERE c.user_id = $1
		ORDER BY c.listing_id
		FOR UPDATE OF c`, buyerID)
	if err != nil {
		return nil, err
	}
	type line struct {
		item       models.CheckoutSessionItem
		priceCents int64
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.item.ListingID, &l.item.Quantity, &l.item.Title, &l.priceCents, &l.item.SellerID); err != nil {
			rows.Close()
			return nil, err
		}
		l.item.Price = float64(l.priceCents) / 100
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if len(lines) == 0 {
		return nil, ErrCartEmpty
	}

	var totalCents int64
	for _, l := range lines {
		if l.item.SellerID == buyerID {
			return nil, ErrCheckoutOwnBuy
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE listings SET stock = stock - $2
			WHERE id = $1 AND archived_at IS NULL AND stock >= $2`, l.item.ListingID, l.item.Quantity)
		if err != nil {
			return nil, err
		}
		if n, err := res.RowsAffected(); err != nil {
			return nil, err
		} else if n == 0 {
			return nil, &models.ListingUnavailableError{ListingID: l.item.ListingID, Title: l.item.Title}
		}
		totalCents += l.priceCents * int64(l.item.Quantity)
	}

	session := &models.CheckoutSession{
		BuyerID: buyerID, Shipping: ship, Total: float64(totalCents) / 100,
		Status: models.CheckoutSessionOpen, ExpiresAt: expiresAt,
	}
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO checkout_sessions (buyer_id, ship_name, ship_phone, ship_line1, ship_line2,
			ship_city, ship_state, ship_postal_code, ship_country, total, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'open', $11)
		RETURNING id, created_at`,
		buyerID, ship.Name, ship.Phone, ship.Line1, ship.Line2, ship.City, ship.State,
		ship.PostalCode, ship.Country, session.Total, expiresAt).Scan(&session.ID, &session.CreatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrLiveCheckoutSession
		}
		return nil, err
	}

	gatewayOrderID, err := createOrder(ctx, totalCents, fmt.Sprintf("checkout_session_%d", session.ID))
	if err != nil {
		return nil, err
	}
	session.GatewayOrderID = gatewayOrderID
	if _, err := tx.ExecContext(ctx,
		`UPDATE checkout_sessions SET gateway_order_id = $2 WHERE id = $1`, session.ID, gatewayOrderID); err != nil {
		return nil, err
	}

	for _, l := range lines {
		it := l.item
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO checkout_session_items (checkout_session_id, listing_id, seller_id, title, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			session.ID, it.ListingID, it.SellerID, it.Title, it.Price, it.Quantity).Scan(&it.ID); err != nil {
			return nil, err
		}
		session.Items = append(session.Items, it)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return session, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
