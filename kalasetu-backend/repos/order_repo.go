package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"kalasetu/models"
)

var (
	ErrCartEmpty      = errors.New("your cart is empty")
	ErrCheckoutOwnBuy = errors.New("you cannot buy your own listing")
	// ErrItemNotCancellable means the Order Item is not in the paid state.
	ErrItemNotCancellable = errors.New("order item cannot be cancelled: only paid items can be")
	// ErrCancelNotCommitted means the refund was accepted but the cancellation
	// did not commit, so the item is still paid and the stock not restocked.
	// Retrying is safe: the refund carries a reference the provider deduplicates.
	ErrCancelNotCommitted = errors.New("order item refunded but the cancellation did not commit")
	// ErrCheckoutSessionNotFound means no Checkout Session matches the
	// gateway order id being confirmed.
	ErrCheckoutSessionNotFound = errors.New("checkout session not found")
	// ErrCheckoutSessionForbidden means the caller is not the Checkout
	// Session's Buyer.
	ErrCheckoutSessionForbidden = errors.New("forbidden: this checkout session is not yours")
	// ErrCheckoutSessionNotConfirmable means the Checkout Session is neither
	// open nor already consumed (it is cancelled or expired), so it cannot
	// be turned into an Order.
	ErrCheckoutSessionNotConfirmable = errors.New("checkout session cannot be confirmed: it is not open")
)

// RefundFunc refunds amountCents of the Order's charge. reference identifies
// the refund so a repeat of one already accepted is a no-op. It runs inside
// the cancellation transaction.
type RefundFunc func(ctx context.Context, chargeID string, amountCents int64, reference string) error

type OrderRepository interface {
	// ConfirmCheckoutSession fulfils the Checkout Session matching
	// gatewayOrderID into an Order: the session is locked, its snapshot
	// becomes the Order and its Items (each paid), the session is marked
	// consumed, and the Buyer's Cart is emptied. It is a single idempotent
	// routine keyed on gatewayOrderID: if the session is already consumed, it
	// returns the Order that resulted the first time rather than creating a
	// second one. It returns ErrCheckoutSessionNotFound if no session matches
	// gatewayOrderID, ErrCheckoutSessionForbidden if it is not buyerID's, and
	// ErrCheckoutSessionNotConfirmable if it is neither open nor already
	// consumed (e.g. cancelled or expired).
	ConfirmCheckoutSession(ctx context.Context, buyerID int, gatewayOrderID, paymentID string) (*models.Order, error)
	// FindByBuyer returns the buyer's orders with their items, newest first, never nil.
	FindByBuyer(ctx context.Context, buyerID int) ([]models.Order, error)
	// FindItemsBySeller returns the Order Items for the seller's Listings, newest
	// first, only those in status when it is non-nil. Never nil.
	FindItemsBySeller(ctx context.Context, sellerID int, status *models.FulfilmentStatus) ([]models.SellerOrderItem, error)
	// FindItem returns the Order Item with its Order's Shipping address, or nil
	// if there is none.
	FindItem(ctx context.Context, id int) (*models.SellerOrderItem, error)
	// AdvanceItem sets the item's status to to only if it is still in from, and
	// reports whether it did, so concurrent updates cannot skip a step.
	AdvanceItem(ctx context.Context, id int, from, to models.FulfilmentStatus) (bool, error)
	// CancelItem, in a single transaction, moves a paid item to cancelled, puts
	// its quantity back into the Listing's stock (Archived or not) and calls
	// refund for the item's price times quantity. Any failure, including
	// refund's (returned unchanged), changes nothing. It returns
	// ErrItemNotCancellable if the item is not paid, so of two concurrent
	// cancels only one refunds, and ErrCancelNotCommitted if the refund was
	// accepted but the transaction did not commit.
	CancelItem(ctx context.Context, id int, refund RefundFunc) error
}

type orderRepository struct {
	db *sql.DB
}

// roundPrice rounds v (a float8 cast of a NUMERIC(12,2) column) to 2 decimal
// places, undoing the imprecision that float8 can introduce.
func roundPrice(v float64) float64 { return math.Round(v*100) / 100 }

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) ConfirmCheckoutSession(ctx context.Context, buyerID int, gatewayOrderID, paymentID string) (*models.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var sessionID, sessBuyerID int
	var status string
	ship := models.ShippingAddress{}
	var total float64
	err = tx.QueryRowContext(ctx, `
		SELECT id, buyer_id, status, ship_name, ship_phone, ship_line1, ship_line2,
			ship_city, ship_state, ship_postal_code, ship_country, total::float8
		FROM checkout_sessions WHERE gateway_order_id = $1 FOR UPDATE`, gatewayOrderID).Scan(
		&sessionID, &sessBuyerID, &status, &ship.Name, &ship.Phone, &ship.Line1, &ship.Line2,
		&ship.City, &ship.State, &ship.PostalCode, &ship.Country, &total)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCheckoutSessionNotFound
	}
	total = roundPrice(total)
	if err != nil {
		return nil, err
	}
	if sessBuyerID != buyerID {
		return nil, ErrCheckoutSessionForbidden
	}

	if status == "consumed" {
		// Already fulfilled, by this call or a concurrent/retried one: replay
		// the Order that resulted the first time instead of making another.
		order, err := r.findOrderByCheckoutSession(ctx, tx, sessionID)
		if err != nil {
			return nil, err
		}
		if order == nil {
			return nil, ErrCheckoutSessionNotConfirmable
		}
		return order, tx.Commit()
	}
	if status != "open" {
		return nil, ErrCheckoutSessionNotConfirmable
	}

	itemRows, err := tx.QueryContext(ctx, `
		SELECT listing_id, seller_id, title, price::float8, quantity
		FROM checkout_session_items WHERE checkout_session_id = $1 ORDER BY id`, sessionID)
	if err != nil {
		return nil, err
	}
	var items []models.OrderItem
	for itemRows.Next() {
		var it models.OrderItem
		if err := itemRows.Scan(&it.ListingID, &it.SellerID, &it.Title, &it.Price, &it.Quantity); err != nil {
			itemRows.Close()
			return nil, err
		}
		it.Price = roundPrice(it.Price)
		it.Status = models.StatusPaid
		items = append(items, it)
	}
	if err := itemRows.Err(); err != nil {
		itemRows.Close()
		return nil, err
	}
	itemRows.Close()

	order := &models.Order{BuyerID: buyerID, Shipping: ship, Total: total}
	// charge_id predates the gateway seam; it now holds paymentID, the
	// gateway's payment identifier, rather than an old PaymentProvider charge id.
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO orders (buyer_id, ship_name, ship_phone, ship_line1, ship_line2,
			ship_city, ship_state, ship_postal_code, ship_country, total, checkout_session_id, charge_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at`,
		buyerID, ship.Name, ship.Phone, ship.Line1, ship.Line2, ship.City, ship.State,
		ship.PostalCode, ship.Country, total, sessionID, paymentID).Scan(&order.ID, &order.CreatedAt); err != nil {
		return nil, err
	}

	for _, it := range items {
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO order_items (order_id, listing_id, seller_id, title, price, quantity, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'paid') RETURNING id`,
			order.ID, it.ListingID, it.SellerID, it.Title, it.Price, it.Quantity).Scan(&it.ID); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, it)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE checkout_sessions SET status = 'consumed' WHERE id = $1`, sessionID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, buyerID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

// findOrderByCheckoutSession returns the Order fulfilled from sessionID, or
// nil if there is none.
func (r *orderRepository) findOrderByCheckoutSession(ctx context.Context, tx *sql.Tx, sessionID int) (*models.Order, error) {
	order := &models.Order{Items: []models.OrderItem{}}
	s := &order.Shipping
	err := tx.QueryRowContext(ctx, `
		SELECT id, buyer_id, ship_name, ship_phone, ship_line1, ship_line2, ship_city, ship_state,
			ship_postal_code, ship_country, total::float8, created_at
		FROM orders WHERE checkout_session_id = $1`, sessionID).Scan(
		&order.ID, &order.BuyerID, &s.Name, &s.Phone, &s.Line1, &s.Line2, &s.City, &s.State,
		&s.PostalCode, &s.Country, &order.Total, &order.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	order.Total = roundPrice(order.Total)

	rows, err := tx.QueryContext(ctx, `
		SELECT id, listing_id, seller_id, title, price::float8, quantity, status
		FROM order_items WHERE order_id = $1 ORDER BY id`, order.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it models.OrderItem
		var status string
		if err := rows.Scan(&it.ID, &it.ListingID, &it.SellerID, &it.Title, &it.Price, &it.Quantity, &status); err != nil {
			return nil, err
		}
		it.Price = roundPrice(it.Price)
		it.Status = models.FulfilmentStatus(strings.ToUpper(status))
		order.Items = append(order.Items, it)
	}
	return order, rows.Err()
}

func (r *orderRepository) FindByBuyer(ctx context.Context, buyerID int) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, ship_name, ship_phone, ship_line1, ship_line2, ship_city, ship_state,
			ship_postal_code, ship_country, total::float8, created_at
		FROM orders WHERE buyer_id = $1 ORDER BY created_at DESC, id DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.Order{}
	index := map[int]int{}
	for rows.Next() {
		o := models.Order{BuyerID: buyerID, Items: []models.OrderItem{}}
		s := &o.Shipping
		if err := rows.Scan(&o.ID, &s.Name, &s.Phone, &s.Line1, &s.Line2, &s.City, &s.State,
			&s.PostalCode, &s.Country, &o.Total, &o.CreatedAt); err != nil {
			return nil, err
		}
		index[o.ID] = len(orders)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return orders, nil
	}

	items, err := r.db.QueryContext(ctx, `
		SELECT i.order_id, i.id, i.listing_id, i.seller_id, i.title, i.price::float8, i.quantity, i.status
		FROM order_items i JOIN orders o ON o.id = i.order_id
		WHERE o.buyer_id = $1 ORDER BY i.id`, buyerID)
	if err != nil {
		return nil, err
	}
	defer items.Close()
	for items.Next() {
		var orderID int
		var it models.OrderItem
		var status string
		if err := items.Scan(&orderID, &it.ID, &it.ListingID, &it.SellerID, &it.Title, &it.Price, &it.Quantity, &status); err != nil {
			return nil, err
		}
		it.Price = math.Round(it.Price*100) / 100
		it.Status = models.FulfilmentStatus(strings.ToUpper(status))
		i := index[orderID]
		orders[i].Items = append(orders[i].Items, it)
	}
	return orders, items.Err()
}

func (r *orderRepository) FindItem(ctx context.Context, id int) (*models.SellerOrderItem, error) {
	items, err := r.queryItems(ctx, `i.id = $1`, id)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *orderRepository) AdvanceItem(ctx context.Context, id int, from, to models.FulfilmentStatus) (bool, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE order_items SET status = $3 WHERE id = $1 AND status = $2`,
		id, strings.ToLower(string(from)), strings.ToLower(string(to)))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *orderRepository) CancelItem(ctx context.Context, id int, refund RefundFunc) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// The row lock serialises concurrent cancels and shipping of the same item.
	var listingID, quantity int
	var priceCents int64
	var status string
	var chargeID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT i.listing_id, i.quantity, (i.price * 100)::bigint, i.status, o.charge_id
		FROM order_items i JOIN orders o ON o.id = i.order_id
		WHERE i.id = $1 FOR UPDATE OF i`, id).Scan(&listingID, &quantity, &priceCents, &status, &chargeID)
	if err != nil {
		return err
	}
	if status != "paid" {
		return ErrItemNotCancellable
	}
	if _, err := tx.ExecContext(ctx, `UPDATE order_items SET status = 'cancelled' WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE listings SET stock = stock + $2 WHERE id = $1`, listingID, quantity); err != nil {
		return err
	}
	if err := refund(ctx, chargeID.String, priceCents*int64(quantity), fmt.Sprintf("item_%d", id)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		// The money is already back with the buyer but nothing else changed.
		// Say so, so a retry is not mistaken for a fresh cancellation.
		return fmt.Errorf("%w: %w", ErrCancelNotCommitted, err)
	}
	return nil
}

func (r *orderRepository) FindItemsBySeller(ctx context.Context, sellerID int, status *models.FulfilmentStatus) ([]models.SellerOrderItem, error) {
	if status == nil {
		return r.queryItems(ctx, `i.seller_id = $1`, sellerID)
	}
	return r.queryItems(ctx, `i.seller_id = $1 AND i.status = $2`, sellerID, strings.ToLower(string(*status)))
}

// queryItems selects Order Items matching where (over aliases i and o), newest first.
func (r *orderRepository) queryItems(ctx context.Context, where string, args ...any) ([]models.SellerOrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT i.id, i.order_id, o.buyer_id, i.listing_id, i.seller_id, i.title, i.price::float8, i.quantity, i.status,
			o.ship_name, o.ship_phone, o.ship_line1, o.ship_line2, o.ship_city, o.ship_state,
			o.ship_postal_code, o.ship_country, o.created_at
		FROM order_items i JOIN orders o ON o.id = i.order_id
		WHERE `+where+`
		ORDER BY o.created_at DESC, i.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.SellerOrderItem{}
	for rows.Next() {
		var it models.SellerOrderItem
		var status string
		s := &it.Shipping
		if err := rows.Scan(&it.ID, &it.OrderID, &it.BuyerID, &it.ListingID, &it.SellerID, &it.Title, &it.Price, &it.Quantity, &status,
			&s.Name, &s.Phone, &s.Line1, &s.Line2, &s.City, &s.State, &s.PostalCode, &s.Country, &it.CreatedAt); err != nil {
			return nil, err
		}
		it.Price = math.Round(it.Price*100) / 100
		it.Status = models.FulfilmentStatus(strings.ToUpper(status))
		items = append(items, it)
	}
	return items, rows.Err()
}
