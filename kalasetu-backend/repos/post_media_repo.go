package repos

import (
	"context"
	"database/sql"
	"kalasetu/models"

	"github.com/lib/pq"
)

// PostMediaRepository is responsible for the post_media table only.
type PostMediaRepository interface {
	// CreateMany inserts a batch of media rows for a single post, preserving
	// upload order via SortOrder.
	CreateMany(ctx context.Context, postID int, media []models.PostMedia) error
	// ListByPost returns all media for a post ordered by their upload order.
	ListByPost(ctx context.Context, postID int) ([]models.PostMedia, error)
	// ListByPosts returns all media for the given posts keyed by post id,
	// ordered by their upload order.
	ListByPosts(ctx context.Context, postIDs []int) (map[int][]models.PostMedia, error)
}

type postMediaRepository struct {
	db *sql.DB
}

func NewPostMediaRepository(db *sql.DB) PostMediaRepository {
	return &postMediaRepository{db: db}
}

const postMediaSelectColumns = `
	pm.id, pm.post_id, pm.object_key, pm.media_type, pm.sort_order, pm.created_at
`

func scanPostMedia(row interface{ Scan(...any) error }) (models.PostMedia, error) {
	var m models.PostMedia
	err := row.Scan(
		&m.ID, &m.PostID, &m.ObjectKey, &m.MediaType, &m.SortOrder, &m.CreatedAt,
	)
	return m, err
}

func (r *postMediaRepository) CreateMany(ctx context.Context, postID int, media []models.PostMedia) error {
	if len(media) == 0 {
		return nil
	}

	query := `
		INSERT INTO post_media (post_id, object_key, media_type, sort_order)
		VALUES ($1, $2, $3, $4)
	`
	for _, m := range media {
		if _, err := r.db.ExecContext(
			ctx, query, postID, m.ObjectKey, m.MediaType, m.SortOrder,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *postMediaRepository) ListByPost(ctx context.Context, postID int) ([]models.PostMedia, error) {
	query := `
		SELECT ` + postMediaSelectColumns + `
		FROM post_media pm
		WHERE pm.post_id = $1
		ORDER BY pm.sort_order ASC, pm.id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	media := []models.PostMedia{}
	for rows.Next() {
		m, err := scanPostMedia(rows)
		if err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, rows.Err()
}

func (r *postMediaRepository) ListByPosts(ctx context.Context, postIDs []int) (map[int][]models.PostMedia, error) {
	if len(postIDs) == 0 {
		return map[int][]models.PostMedia{}, nil
	}

	query := `
		SELECT ` + postMediaSelectColumns + `
		FROM post_media pm
		WHERE pm.post_id = ANY($1)
		ORDER BY pm.sort_order ASC, pm.id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(postIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byPost := make(map[int][]models.PostMedia, len(postIDs))
	for rows.Next() {
		m, err := scanPostMedia(rows)
		if err != nil {
			return nil, err
		}
		byPost[m.PostID] = append(byPost[m.PostID], m)
	}
	return byPost, rows.Err()
}
