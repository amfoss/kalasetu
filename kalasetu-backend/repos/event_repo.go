package repos

import (
	"context"
	"database/sql"
	"errors"
	"kalasetu/models"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) (*models.Event, error)
	FindByID(ctx context.Context, id int) (*models.Event, error)
	List(ctx context.Context) ([]models.Event, error)
	ListByUser(ctx context.Context, userID int) ([]models.Event, error)
	Update(ctx context.Context, id int, input models.EventUpdates) error
	Delete(ctx context.Context, id int) error
}

type eventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepository{db: db}
}

const eventSelectColumns = `
	e.id, e.name, e.start_date::text, e.duration::text, e.host_id, COALESCE(u.name, ''),
	e.banner_key, e.banner_content_type, e.banner_size,
	e.created_at
`

func (r *eventRepository) Create(ctx context.Context, event *models.Event) (*models.Event, error) {
	query := `
		INSERT INTO events (name, start_date, duration, host_id, banner_key, banner_content_type, banner_size)
		VALUES ($1, $2::date, $3::interval, $4, NULLIF($5, ''), $6, NULLIF($7, 0))
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx, query,
		event.Name, event.StartDate, event.Duration, event.HostID,
		event.BannerKey, event.BannerContentType, event.BannerSize,
	).Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		return nil, err
	}
	return event, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEvent(scanner rowScanner) (models.Event, error) {
	var e models.Event
	var bannerKey, bannerContentType sql.NullString
	var bannerSize sql.NullInt64
	err := scanner.Scan(
		&e.ID, &e.Name, &e.StartDate, &e.Duration,
		&e.HostID, &e.HostName,
		&bannerKey, &bannerContentType, &bannerSize,
		&e.CreatedAt,
	)
	if err != nil {
		return models.Event{}, err
	}
	e.BannerKey = bannerKey.String
	e.BannerContentType = bannerContentType.String
	e.BannerSize = bannerSize.Int64
	return e, nil
}

func (r *eventRepository) FindByID(ctx context.Context, id int) (*models.Event, error) {
	query := `
		SELECT ` + eventSelectColumns + `
		FROM events e
		LEFT JOIN users u ON u.id = e.host_id
		WHERE e.id = $1
	`
	event, err := scanEvent(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) List(ctx context.Context) ([]models.Event, error) {
	query := `
		SELECT ` + eventSelectColumns + `
		FROM events e
		LEFT JOIN users u ON u.id = e.host_id
		ORDER BY e.start_date DESC, e.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// Update sets only the fields that were provided (For NULL input, COALESCE keeps the existing value).
func (r *eventRepository) ListByUser(ctx context.Context, userID int) ([]models.Event, error) {
	query := `
		SELECT ` + eventSelectColumns + `
		FROM events e
		LEFT JOIN users u ON u.id = e.host_id
		WHERE e.host_id = $1
		ORDER BY e.start_date DESC, e.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []models.Event{}

	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// Update sets only the fields that were provided (NULL input → COALESCE keeps the existing value).
func (r *eventRepository) Update(ctx context.Context, id int, input models.EventUpdates) error {
	query := `
		UPDATE events
		SET name                 = COALESCE($2, name),
		    start_date           = COALESCE($3::date, start_date),
		    duration             = COALESCE($4::interval, duration),
		    banner_key           = COALESCE($5, banner_key),
		    banner_content_type  = COALESCE($6, banner_content_type),
		    banner_size          = COALESCE($7::bigint, banner_size)
		WHERE id = $1
	`
	var bannerKey, bannerContentType *string
	var bannerSize *int64
	if input.Banner != nil {
		bannerKey = &input.Banner.Key
		bannerContentType = &input.Banner.ContentType
		bannerSize = &input.Banner.Size
	}
	_, err := r.db.ExecContext(
		ctx, query, id,
		input.Name, input.StartDate, input.Duration,
		bannerKey, bannerContentType, bannerSize,
	)
	return err
}

func (r *eventRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, id)
	return err
}
