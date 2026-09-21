package repos

import (
	"context"
	"database/sql"
	"kalasetu/models"
	"strings"
)

type OpportunityRepository interface {
	Create(ctx context.Context, opp *models.Opportunity) error
	Update(ctx context.Context, opp *models.Opportunity) error
	FindByID(ctx context.Context, id int) (*models.Opportunity, error)
	ListByEvent(ctx context.Context, eventID int) ([]models.Opportunity, error)
}

type opportunityRepository struct {
	db *sql.DB
}

func NewOpportunityRepository(db *sql.DB) OpportunityRepository {
	return &opportunityRepository{
		db: db,
	}
}

func (r *opportunityRepository) Create(ctx context.Context, opp *models.Opportunity) error {
	query := `
		INSERT INTO opportunities (event_id, title, description, categories, location, start_date, duration, total_positions, status, host_id)
		VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6, '')::date, CURRENT_DATE), '1 day'::interval, $7, $8, $9)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		opp.EventID, opp.Title, opp.Description, opp.Categories, opp.Location, opp.StartDate, opp.TotalPositions, opp.Status, opp.HostID,
	).Scan(&opp.ID, &opp.CreatedAt)

	return err
}

func (r *opportunityRepository) FindByID(ctx context.Context, id int) (*models.Opportunity, error) {
	query := `
		SELECT id, event_id, COALESCE(title, ''), COALESCE(description, ''), COALESCE(categories, ''), COALESCE(location, ''),
		       TO_CHAR(start_date, 'YYYY-MM-DD'), total_positions, status, host_id, created_at,
		       (SELECT COUNT(*) FROM applications WHERE opportunity_id = opportunities.id) as app_count
		FROM opportunities
		WHERE id = $1
	`
	opp := &models.Opportunity{}
	var startDate sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&opp.ID, &opp.EventID, &opp.Title, &opp.Description, &opp.Categories, &opp.Location,
		&startDate, &opp.TotalPositions, &opp.Status, &opp.HostID, &opp.CreatedAt, &opp.ApplicationsCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if startDate.Valid {
		opp.StartDate = startDate.String
	}
	opp.OpenSlots = opp.TotalPositions - opp.ApplicationsCount
	if opp.OpenSlots < 0 {
		opp.OpenSlots = 0
	}

	return opp, nil
}

func (r *opportunityRepository) ListByEvent(ctx context.Context, eventID int) ([]models.Opportunity, error) {
	query := `
		SELECT id, event_id, COALESCE(title, ''), COALESCE(description, ''), COALESCE(categories, ''), COALESCE(location, ''),
		       TO_CHAR(start_date, 'YYYY-MM-DD'), total_positions, status, host_id, created_at,
		       (SELECT COUNT(*) FROM applications WHERE opportunity_id = opportunities.id) as app_count
		FROM opportunities
		WHERE event_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var opps []models.Opportunity
	for rows.Next() {
		var opp models.Opportunity
		var startDate sql.NullString
		if err := rows.Scan(
			&opp.ID, &opp.EventID, &opp.Title, &opp.Description, &opp.Categories, &opp.Location,
			&startDate, &opp.TotalPositions, &opp.Status, &opp.HostID, &opp.CreatedAt, &opp.ApplicationsCount,
		); err != nil {
			return nil, err
		}
		if startDate.Valid {
			opp.StartDate = startDate.String
		}
		opp.OpenSlots = opp.TotalPositions - opp.ApplicationsCount
		if opp.OpenSlots < 0 {
			opp.OpenSlots = 0
		}
		_ = strings.Split(opp.Categories, ",")
		opps = append(opps, opp)
	}

	return opps, rows.Err()
}

func (r *opportunityRepository) Update(ctx context.Context, opp *models.Opportunity) error {
	query := `
		UPDATE opportunities
		SET title = $1, description = $2, categories = $3, location = $4,
		    start_date = COALESCE(NULLIF($5, '')::date, start_date),
		    total_positions = $6, status = $7
		WHERE id = $8
	`
	_, err := r.db.ExecContext(
		ctx, query,
		opp.Title, opp.Description, opp.Categories, opp.Location,
		opp.StartDate, opp.TotalPositions, opp.Status, opp.ID,
	)
	return err
}

