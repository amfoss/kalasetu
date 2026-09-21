package models

import (
	"io"
	"time"
)

// UploadMedia is the application-level representation of a file uploaded
// through GraphQL multipart. It deliberately mirrors gqlgen's Upload struct
// (reader, original filename, content type) so that services and repositories
// never depend on gqlgen directly.
type UploadMedia struct {
	Reader      io.Reader
	Filename    string
	ContentType string
}

// PostMedia represents one media object attached to a post. The rows are the
// source of truth for a post's media; URL is populated at read time and is not
// persisted.
type PostMedia struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	ObjectKey string    `json:"object_key"`
	MediaType string    `json:"media_type"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url,omitempty"`
}
