package repos

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/lib/pq"
	"kalasetu/models"
)

type ListingRepository interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
	// UserHasAnyRole reports whether the user holds at least one of roles.
	UserHasAnyRole(ctx context.Context, userID int, roles ...string) (bool, error)
	CategoryExists(ctx context.Context, id int) (bool, error)
	// Create stores the listing and its images and returns its id.
	Create(ctx context.Context, sellerID int, input models.CreateListingInput) (int, error)
	// FindByID returns nil, nil when there is no such listing.
	FindByID(ctx context.Context, id int) (*models.Listing, error)
}

type listingRepository struct {
	db *sql.DB
}

func NewListingRepository(db *sql.DB) ListingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) ListCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, category_name FROM categories ORDER BY category_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *listingRepository) Create(ctx context.Context, sellerID int, input models.CreateListingInput) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO listings (seller_id, category_id, title, description, price, stock)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		sellerID, input.CategoryID, input.Title, input.Description,
		strconv.FormatFloat(input.Price, 'f', 2, 64), input.Stock,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	for i, url := range input.ImageURLs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO listing_images (listing_id, position, url) VALUES ($1, $2, $3)`, id, i, url); err != nil {
			return 0, err
		}
	}
	return id, tx.Commit()
}

func (r *listingRepository) FindByID(ctx context.Context, id int) (*models.Listing, error) {
	l := &models.Listing{}
	var location, picture sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT l.id, l.title, l.description, l.price::float8, l.currency, l.stock, l.created_at,
		       c.id, c.category_name,
		       u.id, u.name, u.location, u.profile_picture
		FROM listings l
		JOIN categories c ON c.id = l.category_id
		JOIN users u ON u.id = l.seller_id
		WHERE l.id = $1`, id).Scan(
		&l.ID, &l.Title, &l.Description, &l.Price, &l.Currency, &l.Stock, &l.CreatedAt,
		&l.Category.ID, &l.Category.Name,
		&l.Seller.ID, &l.Seller.Name, &location, &picture,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.Seller.Location = location.String
	l.Seller.ProfilePicture = picture.String

	rows, err := r.db.QueryContext(ctx, `SELECT url FROM listing_images WHERE listing_id = $1 ORDER BY position`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		l.ImageURLs = append(l.ImageURLs, url)
	}
	return l, rows.Err()
}

func (r *listingRepository) UserHasAnyRole(ctx context.Context, userID int, roles ...string) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles ro ON ro.id = ur.role_id
			WHERE ur.user_id = $1 AND ro.role = ANY($2)
		)`, userID, pq.Array(roles)).Scan(&ok)
	return ok, err
}

func (r *listingRepository) CategoryExists(ctx context.Context, id int) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1)`, id).Scan(&ok)
	return ok, err
}
