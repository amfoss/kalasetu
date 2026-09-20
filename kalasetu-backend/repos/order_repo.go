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
)

// PayFunc charges amountCents for reference (the order id) and returns the
// charge id. It runs inside the checkout transaction.
type PayFunc func(ctx context.Context, amountCents int64, reference string) (chargeID string, err error)

type OrderRepository interface {
	// Checkout turns the buyer's Cart into an Order in a single transaction:
	// stock is decremented with a guarded update, pay is called, every item is
	// stored as paid with a snapshot of title and price, and the Cart is emptied.
	// Any failure, including pay's, rolls everything back. It returns
	// ErrCartEmpty, ErrCheckoutOwnBuy or *models.ListingUnavailableError for
	// business failures, and pay's error unchanged.
	Checkout(ctx context.Context, buyerID int, ship models.ShippingAddress, pay PayFunc) (*models.Order, error)
	// FindByBuyer returns the buyer's orders with their items, newest first, never nil.
	FindByBuyer(ctx context.Context, buyerID int) ([]models.Order, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Checkout(ctx context.Context, buyerID int, ship models.ShippingAddress, pay PayFunc) (*models.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

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
		item       models.OrderItem
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
		l.item.Status = models.StatusPaid
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

	order := &models.Order{BuyerID: buyerID, Shipping: ship, Total: float64(totalCents) / 100}
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO orders (buyer_id, ship_name, ship_phone, ship_line1, ship_line2,
			ship_city, ship_state, ship_postal_code, ship_country, total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`,
		buyerID, ship.Name, ship.Phone, ship.Line1, ship.Line2, ship.City, ship.State,
		ship.PostalCode, ship.Country, order.Total).Scan(&order.ID, &order.CreatedAt); err != nil {
		return nil, err
	}

	chargeID, err := pay(ctx, totalCents, fmt.Sprintf("order_%d", order.ID))
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET charge_id = $2 WHERE id = $1`, order.ID, chargeID); err != nil {
		return nil, err
	}

	for _, l := range lines {
		it := l.item
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO order_items (order_id, listing_id, seller_id, title, price, quantity, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'paid') RETURNING id`,
			order.ID, it.ListingID, it.SellerID, it.Title, it.Price, it.Quantity).Scan(&it.ID); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, it)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE user_id = $1`, buyerID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
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
