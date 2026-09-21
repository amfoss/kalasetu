package repos

import (
	"context"
	"database/sql"
	"errors"
	"kalasetu/models"
)

type ProfileRepository interface {
	GetBasics(ctx context.Context, userID int) (*models.Profile, error)
	CountArtworks(ctx context.Context, userID int) (int, error)
	ListArtworkImages(ctx context.Context, userID int) ([]string, error)
	CountTotalLikes(ctx context.Context, userID int) (int, error)
	CountFollowers(ctx context.Context, userID int) (int, error)
	CountFollowing(ctx context.Context, userID int) (int, error)
	ListSkills(ctx context.Context, userID int) ([]string, error)
	ListAchievements(ctx context.Context, userID int) ([]models.Achievement, error)
	ListRecentPosts(ctx context.Context, userID int, limit int) ([]models.ProfilePost, error)
}

type profileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) GetBasics(ctx context.Context, userID int) (*models.Profile, error) {
	query := `
		SELECT id, name, COALESCE(location, ''),
		       COALESCE(bio, ''), COALESCE(profile_picture, ''), email
		FROM users
		WHERE id = $1
	`
	profile := &models.Profile{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.Name,
		&profile.Location,
		&profile.Bio,
		&profile.ProfilePicture,
		&profile.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return profile, nil
}

func (r *profileRepository) CountArtworks(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM posts WHERE user_id = $1
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *profileRepository) ListArtworkImages(ctx context.Context, userID int) ([]string, error) {
	query := `
		SELECT pm.object_key
		FROM post_media pm
		JOIN posts p ON p.id = pm.post_id
		WHERE p.user_id = $1
		ORDER BY pm.sort_order ASC, pm.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []string{}
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, err
		}
		images = append(images, uri)
	}
	return images, rows.Err()
}

func (r *profileRepository) CountTotalLikes(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM likes l
		JOIN posts p ON l.parent_id = p.id AND l.parent_type = 'post'
		WHERE p.user_id = $1
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *profileRepository) CountFollowers(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM user_follows WHERE following_id = $1
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *profileRepository) CountFollowing(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*) FROM user_follows WHERE follower_id = $1
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *profileRepository) ListSkills(ctx context.Context, userID int) ([]string, error) {
	query := `
		SELECT l.label_name
		FROM user_labels ul
		JOIN labels l ON l.id = ul.label_id
		WHERE ul.user_id = $1
		ORDER BY l.label_name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []string{}
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

func (r *profileRepository) ListAchievements(ctx context.Context, userID int) ([]models.Achievement, error) {
	query := `
		SELECT title, description, icon_type
		FROM achievements
		WHERE user_id = $1
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	achievements := []models.Achievement{}
	for rows.Next() {
		var a models.Achievement
		if err := rows.Scan(&a.Title, &a.Description, &a.IconType); err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}
	return achievements, rows.Err()
}

func (r *profileRepository) ListRecentPosts(ctx context.Context, userID int, limit int) ([]models.ProfilePost, error) {
	query := `
		SELECT p.id, p.content,
		       COALESCE((SELECT pm.media_type FROM post_media pm WHERE pm.post_id = p.id ORDER BY pm.sort_order ASC LIMIT 1), ''),
		       COALESCE((SELECT pm.object_key  FROM post_media pm WHERE pm.post_id = p.id ORDER BY pm.sort_order ASC LIMIT 1), ''),
		       p.created_at,
		       (SELECT COUNT(*) FROM likes    l WHERE l.parent_type = 'post' AND l.parent_id = p.id) AS like_count,
		       (SELECT COUNT(*) FROM comments c WHERE c.parent_type = 'post' AND c.parent_id = p.id) AS comment_count
		FROM posts p
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.ProfilePost{}
	for rows.Next() {
		var p models.ProfilePost
		if err := rows.Scan(
			&p.ID,
			&p.Content,
			&p.MediaType,
			&p.MediaURI,
			&p.CreatedAt,
			&p.LikeCount,
			&p.CommentCount,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}