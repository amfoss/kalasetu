package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type OTPRepository interface {
	CreateOTP(ctx context.Context, email, otp, purpose string, expiresAt time.Time) error
	VerifyAndConsumeOTP(ctx context.Context, email, otp, purpose string) (bool, error)
	IsEmailVerified(ctx context.Context, email, purpose string, withinDuration time.Duration) (bool, error)
}

type otpRepository struct {
	db *sql.DB
}

func NewOTPRepository(db *sql.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) CreateOTP(ctx context.Context, email, otp, purpose string, expiresAt time.Time) error {
	// Delete any existing unused OTPs for the same email & purpose
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM email_otps WHERE email = $1 AND purpose = $2 AND used = false`,
		email, purpose,
	)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO email_otps (email, otp, purpose, expires_at) VALUES ($1, $2, $3, $4)`,
		email, otp, purpose, expiresAt,
	)
	return err
}

func (r *otpRepository) VerifyAndConsumeOTP(ctx context.Context, email, otp, purpose string) (bool, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM email_otps
		WHERE email = $1 AND otp = $2 AND purpose = $3 AND used = false AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
		LIMIT 1
	`, email, otp, purpose).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Mark as used
	_, err = r.db.ExecContext(ctx, `UPDATE email_otps SET used = true WHERE id = $1`, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *otpRepository) IsEmailVerified(ctx context.Context, email, purpose string, withinDuration time.Duration) (bool, error) {
	since := time.Now().Add(-withinDuration)
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM email_otps
			WHERE email = $1 AND purpose = $2 AND used = true AND created_at >= $3
		)
	`, email, purpose, since).Scan(&exists)

	return exists, err
}
