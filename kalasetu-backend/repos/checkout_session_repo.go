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

// ErrNoOpenCheckoutSession means the Buyer has no open Checkout Session to
// cancel.
var ErrNoOpenCheckoutSession = errors.New("you have no checkout in progress")

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
	// Cancel cancels buyerID's open Checkout Session, releasing its reserved
	// Stock immediately, whether or not the Listing has since been Archived,
	// so the Buyer can start a new Checkout Session right away. It returns
	// ErrNoOpenCheckoutSession if the Buyer has none open.
	Cancel(ctx context.Context, buyerID int) error
	// ReleaseExpired releases every open Checkout Session whose expiry has
	// already passed, restoring each one's reserved Stock. It is the
	// background sweeper's entry point.
	ReleaseExpired(ctx context.Context) error
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

	// Release the Buyer's own expired session, if any, before checking for
	// a live one: an abandoned checkout must not block a fresh one.
	if err := releaseExpiredSessionsTx(ctx, tx, &buyerID); err != nil {
		return nil, err
	}

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

	// Release, before contending on their Stock, any of these Listings still
	// held by another Buyer's session whose expiry has passed: an abandoned
	// checkout must not make a Listing look unavailable to everyone else
	// until the sweeper next runs.
	listingIDs := make([]int, len(lines))
	for i, l := range lines {
		listingIDs[i] = l.item.ListingID
	}
	if err := releaseExpiredSessionsForListingsTx(ctx, tx, listingIDs); err != nil {
		return nil, err
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

func (r *checkoutSessionRepository) Cancel(ctx context.Context, buyerID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var sessionID int
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM checkout_sessions WHERE buyer_id = $1 AND status = 'open' FOR UPDATE`, buyerID).Scan(&sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoOpenCheckoutSession
	}
	if err != nil {
		return err
	}

	if err := releaseSessionTx(ctx, tx, sessionID, "cancelled"); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *checkoutSessionRepository) ReleaseExpired(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := releaseExpiredSessionsTx(ctx, tx, nil); err != nil {
		return err
	}
	return tx.Commit()
}

// restoreCheckoutSessionStockTx restores sessionID's reserved Stock to each
// Listing it reserved, whether or not the Listing has since been Archived.
func restoreCheckoutSessionStockTx(ctx context.Context, tx *sql.Tx, sessionID int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE listings l SET stock = l.stock + i.quantity
		FROM checkout_session_items i
		WHERE i.checkout_session_id = $1 AND i.listing_id = l.id`, sessionID)
	return err
}

// releaseSessionTx releases sessionID: its reserved Stock is restored (see
// restoreCheckoutSessionStockTx) and its status set to newStatus.
func releaseSessionTx(ctx context.Context, tx *sql.Tx, sessionID int, newStatus string) error {
	if err := restoreCheckoutSessionStockTx(ctx, tx, sessionID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE checkout_sessions SET status = $2 WHERE id = $1`, sessionID, newStatus)
	return err
}

// findExpiredSessionIDsTx locks and returns the ids of every open Checkout
// Session whose expiry has passed as of now, matching where (over the
// checkout_sessions row, aliased s).
func findExpiredSessionIDsTx(ctx context.Context, tx *sql.Tx, where string, args ...any) ([]int, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.id FROM checkout_sessions s
		WHERE s.status = 'open' AND s.expires_at <= now() AND `+where+`
		FOR UPDATE OF s`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// releaseExpiredSessionsTx expires every open Checkout Session whose expiry
// has passed as of now (buyerID's alone, when given), restoring each one's
// reserved Stock. It runs inside tx, first in any transaction that is about
// to contend on Stock, so an abandoned session's Stock is never
// permanently lost.
func releaseExpiredSessionsTx(ctx context.Context, tx *sql.Tx, buyerID *int) error {
	where := `true`
	var args []any
	if buyerID != nil {
		where = `s.buyer_id = $1`
		args = append(args, *buyerID)
	}
	ids, err := findExpiredSessionIDsTx(ctx, tx, where, args...)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := releaseSessionTx(ctx, tx, id, "expired"); err != nil {
			return err
		}
	}
	return nil
}

// releaseExpiredSessionsForListingsTx expires every open Checkout Session
// whose expiry has passed and which reserved Stock for one of listingIDs,
// restoring each one's reserved Stock. It runs immediately before the
// contended Stock update on those Listings, so a Listing held by another
// Buyer's abandoned session becomes available the moment it expires rather
// than waiting for the background sweeper.
func releaseExpiredSessionsForListingsTx(ctx context.Context, tx *sql.Tx, listingIDs []int) error {
	if len(listingIDs) == 0 {
		return nil
	}
	ids, err := findExpiredSessionIDsTx(ctx, tx,
		`s.id IN (SELECT DISTINCT checkout_session_id FROM checkout_session_items WHERE listing_id = ANY($1))`,
		pq.Array(listingIDs))
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := releaseSessionTx(ctx, tx, id, "expired"); err != nil {
			return err
		}
	}
	return nil
}
