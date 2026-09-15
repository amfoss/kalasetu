package models

import "time"

type Event struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"` // "2006-01-02"
	Duration  string    `json:"duration"`   // Postgres interval literal, e.g. "2 days"
	HostID    *int      `json:"host_id"`
	HostName  string    `json:"host_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Banner metadata. Only the S3 object key is persisted; BannerURL is
	// resolved at read time and is never stored in PostgreSQL.
	BannerKey         string `json:"banner_key"`
	BannerContentType string `json:"banner_content_type"`
	BannerSize        int64  `json:"banner_size"`
	BannerURL         string `json:"banner_url,omitempty"`
}

type CreateEventInput struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	Duration  string `json:"duration" binding:"required"`
	Banner    *UploadMedia
}

type UpdateEventInput struct {
	Name      *string `json:"name"`
	StartDate *string `json:"start_date"`
	Duration  *string `json:"duration"`
	Banner    *UploadMedia
}

// EventBanner is the persisted metadata of an event's banner image. The S3
// object key is the source of truth; the public URL is derived from it at read
// time.
type EventBanner struct {
	Key         string
	ContentType string
	Size        int64
}

// EventUpdates is the repository-level write shape for updating an event. Nil
// fields are left unchanged (COALESCE), so a non-nil Banner replaces the
// existing one while a nil Banner keeps it untouched.
type EventUpdates struct {
	Name      *string
	StartDate *string
	Duration  *string
	Banner    *EventBanner
}
