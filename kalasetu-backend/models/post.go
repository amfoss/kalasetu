package models

import "time"

type Post struct {
	ID           int         `json:"id"`
	UserID       int         `json:"user_id"`
	UserName     string      `json:"user_name"`
	Content      string      `json:"content"`
	Media        []PostMedia `json:"media"`
	CategoryID   *int        `json:"category_id"`
	CategoryName string      `json:"category_name"`
	LikeCount    int         `json:"like_count"`
	CommentCount int         `json:"comment_count"`
	IsLikedByMe  bool        `json:"is_liked_by_me"`
	CreatedAt    time.Time   `json:"created_at"`
}

type CreatePostInput struct {
	Content    string        `json:"content"`
	Media      []UploadMedia `json:"media"`
	CategoryID *int          `json:"category_id"`
}

type UpdatePostInput struct {
	Content    *string `json:"content"`
	CategoryID *int    `json:"category_id"`
}

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCommentInput struct {
	PostID  int    `json:"post_id"`
	Content string `json:"content"`
}

type UpdateCommentInput struct {
	Content string `json:"content"`
}
