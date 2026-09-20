package repos

import (
	"context"
	"database/sql"
	"errors"

	"kalasetu/models"
)

type CartRepository interface {
	// Items returns the user's lines in the order they were first added, never nil.
	Items(ctx context.Context, userID int) ([]models.CartItem, error)
	// Quantity returns 0 when the listing is not in the user's cart.
	Quantity(ctx context.Context, userID, listingID int) (int, error)
	// Set stores the quantity (> 0), creating the line if needed.
	Set(ctx context.Context, userID, listingID, quantity int) error
	// Remove deletes the line; removing an absent line is a no-op.
	Remove(ctx context.Context, userID, listingID int) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) Items(ctx context.Context, userID int) ([]models.CartItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT listing_id, quantity FROM cart_items WHERE user_id = $1 ORDER BY created_at, listing_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.CartItem{}
	for rows.Next() {
		var it models.CartItem
		if err := rows.Scan(&it.ListingID, &it.Quantity); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *cartRepository) Quantity(ctx context.Context, userID, listingID int) (int, error) {
	var q int
	err := r.db.QueryRowContext(ctx,
		`SELECT quantity FROM cart_items WHERE user_id = $1 AND listing_id = $2`, userID, listingID).Scan(&q)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return q, err
}

func (r *cartRepository) Set(ctx context.Context, userID, listingID, quantity int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_items (user_id, listing_id, quantity) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, listing_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
		userID, listingID, quantity)
	return err
}

func (r *cartRepository) Remove(ctx context.Context, userID, listingID int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items WHERE user_id = $1 AND listing_id = $2`, userID, listingID)
	return err
}
