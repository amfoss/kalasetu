package repos

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

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
	// FindByIDs returns the listings that exist, archived ones included, in no
	// particular order, never nil.
	FindByIDs(ctx context.Context, ids []int) ([]models.Listing, error)
	// FindBySeller returns all of the seller's listings, archived ones included,
	// newest first, never nil.
	FindBySeller(ctx context.Context, sellerID int) ([]models.Listing, error)
	// Update applies the non-nil fields of input to a non-archived listing and
	// reports whether it did; false means the listing is archived or gone.
	Update(ctx context.Context, id int, input models.UpdateListingInput) (bool, error)
	// Archive marks the listing archived; archiving twice keeps the first time.
	Archive(ctx context.Context, id int) error
	// Search returns non-archived listings matching q, never nil.
	Search(ctx context.Context, q models.ListingQuery) ([]models.Listing, error)
	// Featured returns up to limit non-archived, in-stock listings in random
	// order, never nil.
	Featured(ctx context.Context, limit int) ([]models.Listing, error)
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
		SELECT l.id, l.title, l.description, l.price::float8, l.currency, l.stock, l.created_at, l.archived_at IS NOT NULL,
		       c.id, c.category_name,
		       u.id, u.name, u.location, u.profile_picture
		FROM listings l
		JOIN categories c ON c.id = l.category_id
		JOIN users u ON u.id = l.seller_id
		WHERE l.id = $1`, id).Scan(
		&l.ID, &l.Title, &l.Description, &l.Price, &l.Currency, &l.Stock, &l.CreatedAt, &l.Archived,
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

func (r *listingRepository) FindByIDs(ctx context.Context, ids []int) ([]models.Listing, error) {
	ids64 := make([]int64, len(ids))
	for i, id := range ids {
		ids64[i] = int64(id)
	}
	return r.queryListings(ctx, listingSelect+" WHERE l.id = ANY($1)", pq.Array(ids64))
}

func (r *listingRepository) FindBySeller(ctx context.Context, sellerID int) ([]models.Listing, error) {
	return r.queryListings(ctx, listingSelect+`
		WHERE l.seller_id = $1
		ORDER BY l.created_at DESC, l.id DESC`, sellerID)
}

// Update writes only the fields being changed, so a seller editing a title
// cannot overwrite a stock count changed meanwhile.
func (r *listingRepository) Update(ctx context.Context, id int, input models.UpdateListingInput) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var lockedID int
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM listings WHERE id = $1 AND archived_at IS NULL FOR UPDATE`, id).Scan(&lockedID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	var sets []string
	var args []any
	set := func(column string, v any) {
		args = append(args, v)
		sets = append(sets, column+" = $"+strconv.Itoa(len(args)))
	}
	if input.Title != nil {
		set("title", *input.Title)
	}
	if input.Description != nil {
		set("description", *input.Description)
	}
	if input.Price != nil {
		set("price", strconv.FormatFloat(*input.Price, 'f', 2, 64))
	}
	if input.Stock != nil {
		set("stock", *input.Stock)
	}
	if input.CategoryID != nil {
		set("category_id", *input.CategoryID)
	}
	if len(sets) > 0 {
		args = append(args, id)
		if _, err := tx.ExecContext(ctx,
			"UPDATE listings SET "+strings.Join(sets, ", ")+" WHERE id = $"+strconv.Itoa(len(args)), args...); err != nil {
			return false, err
		}
	}

	if input.ImageURLs != nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM listing_images WHERE listing_id = $1`, id); err != nil {
			return false, err
		}
		for i, url := range input.ImageURLs {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO listing_images (listing_id, position, url) VALUES ($1, $2, $3)`, id, i, url); err != nil {
				return false, err
			}
		}
	}
	return true, tx.Commit()
}

func (r *listingRepository) Archive(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE listings SET archived_at = COALESCE(archived_at, now()) WHERE id = $1`, id)
	return err
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

const listingSelect = `
	SELECT l.id, l.title, l.description, l.price::float8, l.currency, l.stock, l.created_at, l.archived_at IS NOT NULL,
	       c.id, c.category_name,
	       u.id, u.name, u.location, u.profile_picture
	FROM listings l
	JOIN categories c ON c.id = l.category_id
	JOIN users u ON u.id = l.seller_id`

// likeEscaper makes user text literal inside a LIKE pattern (ESCAPE '\').
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (r *listingRepository) Search(ctx context.Context, q models.ListingQuery) ([]models.Listing, error) {
	where := []string{"l.archived_at IS NULL"}
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	if q.Query != "" {
		p := arg("%" + likeEscaper.Replace(q.Query) + "%")
		where = append(where, "(l.title ILIKE "+p+` ESCAPE '\' OR l.description ILIKE `+p+` ESCAPE '\')`)
	}
	if q.CategoryID != nil {
		where = append(where, "l.category_id = "+arg(*q.CategoryID))
	}
	if q.MinPrice != nil {
		where = append(where, "l.price >= "+arg(*q.MinPrice)+"::numeric")
	}
	if q.MaxPrice != nil {
		where = append(where, "l.price <= "+arg(*q.MaxPrice)+"::numeric")
	}
	if q.InStockOnly {
		where = append(where, "l.stock > 0")
	}

	order := "l.created_at DESC, l.id DESC"
	switch q.Sort {
	case models.SortPriceAsc:
		order = "l.price ASC, l.id DESC"
	case models.SortPriceDesc:
		order = "l.price DESC, l.id DESC"
	}

	query := listingSelect + " WHERE " + strings.Join(where, " AND ") +
		" ORDER BY " + order + " LIMIT " + arg(q.Limit) + " OFFSET " + arg(q.Offset)
	return r.queryListings(ctx, query, args...)
}

func (r *listingRepository) Featured(ctx context.Context, limit int) ([]models.Listing, error) {
	return r.queryListings(ctx, listingSelect+`
		WHERE l.archived_at IS NULL AND l.stock > 0
		ORDER BY random() LIMIT $1`, limit)
}

// queryListings runs a listingSelect query and attaches images with one extra
// query rather than one per listing.
func (r *listingRepository) queryListings(ctx context.Context, query string, args ...any) ([]models.Listing, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := []models.Listing{}
	for rows.Next() {
		var l models.Listing
		var location, picture sql.NullString
		if err := rows.Scan(
			&l.ID, &l.Title, &l.Description, &l.Price, &l.Currency, &l.Stock, &l.CreatedAt, &l.Archived,
			&l.Category.ID, &l.Category.Name,
			&l.Seller.ID, &l.Seller.Name, &location, &picture,
		); err != nil {
			return nil, err
		}
		l.Seller.Location = location.String
		l.Seller.ProfilePicture = picture.String
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(listings) == 0 {
		return listings, nil
	}

	ids := make([]int64, len(listings))
	index := make(map[int]int, len(listings))
	for i, l := range listings {
		ids[i] = int64(l.ID)
		index[l.ID] = i
	}
	imgRows, err := r.db.QueryContext(ctx,
		`SELECT listing_id, url FROM listing_images WHERE listing_id = ANY($1) ORDER BY listing_id, position`,
		pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer imgRows.Close()
	for imgRows.Next() {
		var id int
		var url string
		if err := imgRows.Scan(&id, &url); err != nil {
			return nil, err
		}
		i := index[id]
		listings[i].ImageURLs = append(listings[i].ImageURLs, url)
	}
	return listings, imgRows.Err()
}
