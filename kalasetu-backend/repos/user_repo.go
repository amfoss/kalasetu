package repos

import (
	"context"
	"database/sql"
	"errors"
	"kalasetu/models"
	"time"
	"fmt"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	StartOnboarding(ctx context.Context, userID int, onboardingUser models.OnboardingUser) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (email, password, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		user.Email, user.Password, user.Name, now, now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or custom error like ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) StartOnboarding(ctx context.Context, userID int, onboardingUser models.OnboardingUser) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update user
	query := `
		UPDATE users 
		SET name = $1, location = $2, bio = $3, profile_picture = $4, updated_at = $5
		WHERE id = $6
	`
	_, err = tx.ExecContext(ctx, query, onboardingUser.Name, onboardingUser.Location, onboardingUser.Bio, onboardingUser.ProfilePicture, time.Now(), userID)
	if err != nil {
		return err
	}

	// 2. Get role and link user_roles
	query = `
		SELECT id FROM roles WHERE role = $1
	`
	var roleID int
	err = tx.QueryRowContext(ctx, query, onboardingUser.Role).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("role lookup failed for %q: %w", onboardingUser.Role, err)
	}

	query = `
		INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return err
	}

	// 3. Insert labels and link user_labels
	for _, labelName := range onboardingUser.Labels {
		query = `
			INSERT INTO labels (label_name) VALUES ($1)
			ON CONFLICT (label_name) DO UPDATE SET label_name = EXCLUDED.label_name
			RETURNING id
		`
		var labelID int
		err = tx.QueryRowContext(ctx, query, labelName).Scan(&labelID)
		if err != nil {
			return fmt.Errorf("label lookup/insert failed for %q: %w", labelName, err)
		}

		query = `
			INSERT INTO user_labels (user_id, label_id) VALUES ($1, $2)
			ON CONFLICT (user_id, label_id) DO NOTHING
		`
		_, err = tx.ExecContext(ctx, query, userID, labelID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
