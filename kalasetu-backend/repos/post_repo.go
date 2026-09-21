package repos

import (
	"context"
	"database/sql"
	"errors"
	"kalasetu/models"
)

// postParentType identifies posts in the polymorphic likes/comments tables.
const postParentType = "post"

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) (*models.Post, error)
	FindByID(ctx context.Context, id int, currentUserID int) (*models.Post, error)
	List(ctx context.Context, currentUserID int, limit, offset int) ([]models.Post, error)
	ListByUser(ctx context.Context, authorUserID int, currentUserID int, limit, offset int) ([]models.Post, error)
	Update(ctx context.Context, id int, input models.UpdatePostInput) error
	Delete(ctx context.Context, id int) error
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

const postSelectColumns = `
	p.id, p.user_id, COALESCE(u.name, ''), p.content,
	p.category_id, COALESCE(cat.category_name, ''),
	(SELECT COUNT(*) FROM likes l WHERE l.parent_type = 'post' AND l.parent_id = p.id),
	(SELECT COUNT(*) FROM comments cm WHERE cm.parent_type = 'post' AND cm.parent_id = p.id),
	EXISTS(SELECT 1 FROM likes l WHERE l.parent_type = 'post' AND l.parent_id = p.id AND l.user_id = $2),
	p.created_at
`

func (r *postRepository) Create(ctx context.Context, post *models.Post) (*models.Post, error) {
	query := `
		INSERT INTO posts (user_id, content, category_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx, query,
		post.UserID, post.Content, post.CategoryID,
	).Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *postRepository) FindByID(ctx context.Context, id int, currentUserID int) (*models.Post, error) {
	query := `
		SELECT ` + postSelectColumns + `
		FROM posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN categories cat ON cat.id = p.category_id
		WHERE p.id = $1
	`
	post := &models.Post{}
	err := r.db.QueryRowContext(ctx, query, id, currentUserID).Scan(
		&post.ID, &post.UserID, &post.UserName, &post.Content,
		&post.CategoryID, &post.CategoryName,
		&post.LikeCount, &post.CommentCount, &post.IsLikedByMe, &post.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return post, nil
}

func (r *postRepository) List(ctx context.Context, currentUserID int, limit, offset int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT ` + postSelectColumns + `
		FROM posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN categories cat ON cat.id = p.category_id
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $1 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, limit, currentUserID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.UserName, &p.Content,
			&p.CategoryID, &p.CategoryName,
			&p.LikeCount, &p.CommentCount, &p.IsLikedByMe, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) ListByUser(ctx context.Context, authorUserID int, currentUserID int, limit, offset int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT ` + postSelectColumns + `
		FROM posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN categories cat ON cat.id = p.category_id
		WHERE p.user_id = $3
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $1 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, limit, currentUserID, authorUserID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.UserName, &p.Content,
			&p.CategoryID, &p.CategoryName,
			&p.LikeCount, &p.CommentCount, &p.IsLikedByMe, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// Update sets only the fields that were provided (NULL input → COALESCE keeps the existing value).
func (r *postRepository) Update(ctx context.Context, id int, input models.UpdatePostInput) error {
	query := `
		UPDATE posts
		SET content     = COALESCE($2, content),
		    category_id = COALESCE($3, category_id)
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, input.Content, input.CategoryID)
	return err
}

func (r *postRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM likes WHERE parent_type = 'post' AND parent_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE parent_type = 'post' AND parent_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id); err != nil {
		return err
	}

	return tx.Commit()
}

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) (*models.Comment, error)
	FindByID(ctx context.Context, id int) (*models.Comment, error)
	Update(ctx context.Context, id int, content string) error
	Delete(ctx context.Context, id int) error
	ListByPost(ctx context.Context, postID int) ([]models.Comment, error)
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

const commentSelectColumns = `
	c.id, c.parent_id, c.user_id, COALESCE(u.name, ''), c.content, c.created_at
`

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	query := `
		INSERT INTO comments (user_id, parent_type, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx, query,
		comment.UserID, postParentType, comment.PostID, comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *commentRepository) FindByID(ctx context.Context, id int) (*models.Comment, error) {
	query := `
		SELECT ` + commentSelectColumns + `
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1
	`
	comment := &models.Comment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID, &comment.PostID, &comment.UserID, &comment.UserName,
		&comment.Content, &comment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return comment, nil
}

func (r *commentRepository) Update(ctx context.Context, id int, content string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE comments SET content = $1 WHERE id = $2`, content, id)
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	return err
}

func (r *commentRepository) ListByPost(ctx context.Context, postID int) ([]models.Comment, error) {
	query := `
		SELECT ` + commentSelectColumns + `
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.parent_type = $1 AND c.parent_id = $2
		ORDER BY c.created_at ASC, c.id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, postParentType, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(
			&c.ID, &c.PostID, &c.UserID, &c.UserName,
			&c.Content, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

type LikeRepository interface {
	Like(ctx context.Context, userID, postID int) error
	Unlike(ctx context.Context, userID, postID int) error
	ListUsersByPost(ctx context.Context, postID int) ([]models.Author, error)
}

type likeRepository struct {
	db *sql.DB
}

func NewLikeRepository(db *sql.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Like(ctx context.Context, userID, postID int) error {
	query := `
		INSERT INTO likes (user_id, parent_type, parent_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, parent_type, parent_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, userID, postParentType, postID)
	return err
}

func (r *likeRepository) Unlike(ctx context.Context, userID, postID int) error {
	query := `
		DELETE FROM likes
		WHERE user_id = $1 AND parent_type = $2 AND parent_id = $3
	`
	_, err := r.db.ExecContext(ctx, query, userID, postParentType, postID)
	return err
}

func (r *likeRepository) ListUsersByPost(ctx context.Context, postID int) ([]models.Author, error) {
	query := `
		SELECT u.id, u.name
		FROM likes l
		JOIN users u ON u.id = l.user_id
		WHERE l.parent_type = $1 AND l.parent_id = $2
		ORDER BY l.created_at DESC, l.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, postParentType, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	authors := []models.Author{}
	for rows.Next() {
		var a models.Author
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, rows.Err()
}
