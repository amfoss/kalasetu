package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"kalasetu/models"
	"kalasetu/repos"
	"kalasetu/storage"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrForbidden         = errors.New("you are not the host of this event")
	ErrInvalidBannerType = errors.New("invalid banner type: must be image/jpeg, image/png or image/webp")
	ErrBannerTooLarge    = errors.New("banner is too large: maximum size is 5 MB")
)

const maxEventBannerSize = 5 * 1024 * 1024

var allowedBannerContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type EventService interface {
	Create(ctx context.Context, userID int, input models.CreateEventInput) (*models.Event, error)
	List(ctx context.Context) ([]models.Event, error)
	ListByUser(ctx context.Context, userID int) ([]models.Event, error)
	GetByID(ctx context.Context, id int) (*models.Event, error)
	Update(ctx context.Context, userID, id int, input models.UpdateEventInput) (*models.Event, error)
	Delete(ctx context.Context, userID, id int) error
}

type eventService struct {
	eventRepo repos.EventRepository
	storage   storage.ObjectStorage
}

func NewEventService(eventRepo repos.EventRepository, objectStorage storage.ObjectStorage) EventService {
	return &eventService{
		eventRepo: eventRepo,
		storage:   objectStorage,
	}
}

func (s *eventService) Create(ctx context.Context, userID int, input models.CreateEventInput) (*models.Event, error) {
	event, err := s.eventRepo.Create(ctx, &models.Event{
		Name:      input.Name,
		StartDate: input.StartDate,
		Duration:  input.Duration,
		HostID:    &userID,
	})
	if err != nil {
		return nil, err
	}

	if input.Banner != nil {
		banner, err := s.uploadBanner(ctx, event.ID, input.Banner)
		if err != nil {
			return nil, err
		}
		if err := s.eventRepo.Update(ctx, event.ID, models.EventUpdates{Banner: banner}); err != nil {
			// The banner reached object storage but its metadata could not be
			// recorded; remove the object so nothing is left orphaned.
			s.cleanupBanner(ctx, banner.Key)
			return nil, err
		}
	}

	// Refetch so the response includes host_name and DB-normalized fields.
	return s.GetByID(ctx, event.ID)
}

func (s *eventService) List(ctx context.Context) ([]models.Event, error) {
	events, err := s.eventRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateEvents(ctx, events); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *eventService) ListByUser(ctx context.Context, userID int) ([]models.Event, error) {
	events, err := s.eventRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateEvents(ctx, events); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *eventService) GetByID(ctx context.Context, id int) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}
	if err := s.hydrateBanner(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *eventService) Update(ctx context.Context, userID, id int, input models.UpdateEventInput) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}
	if event.HostID == nil || *event.HostID != userID {
		return nil, ErrForbidden
	}

	updates := models.EventUpdates{
		Name:      input.Name,
		StartDate: input.StartDate,
		Duration:  input.Duration,
	}

	if input.Banner == nil {
		if err := s.eventRepo.Update(ctx, id, updates); err != nil {
			return nil, err
		}
		return s.GetByID(ctx, id)
	}

	// Banner replacement: upload the new object first, then persist its
	// metadata, and only then remove the previous object.
	banner, err := s.uploadBanner(ctx, id, input.Banner)
	if err != nil {
		return nil, err
	}
	updates.Banner = banner
	if err := s.eventRepo.Update(ctx, id, updates); err != nil {
		// The new banner is stored but could not be recorded; clean it up so the
		// previous (still referenced) banner remains intact.
		s.cleanupBanner(ctx, banner.Key)
		return nil, err
	}
	if event.BannerKey != "" {
		// Failures here must not destroy the new banner, so deletion is best-effort.
		s.cleanupBanner(ctx, event.BannerKey)
	}
	return s.GetByID(ctx, id)
}

func (s *eventService) Delete(ctx context.Context, userID, id int) error {
	event, err := s.eventRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return ErrEventNotFound
	}
	if event.HostID == nil || *event.HostID != userID {
		return ErrForbidden
	}

	if err := s.eventRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Best-effort removal of the underlying object; the event row is already gone.
	if event.BannerKey != "" {
		s.cleanupBanner(ctx, event.BannerKey)
	}
	return nil
}

// uploadBanner validates and buffers the incoming banner, uploads it to object
// storage under a fresh, storage-friendly key and returns its metadata.
func (s *eventService) uploadBanner(ctx context.Context, eventID int, upload *models.UploadMedia) (*models.EventBanner, error) {
	if upload == nil {
		return nil, nil
	}
	if s.storage == nil {
		return nil, ErrStorageUnconfigured
	}

	buf, err := s.prepareBanner(upload)
	if err != nil {
		return nil, err
	}

	key, err := s.generateBannerKey(eventID, upload.Filename)
	if err != nil {
		return nil, err
	}
	if err := s.storage.Upload(ctx, key, bytes.NewReader(buf), upload.ContentType); err != nil {
		return nil, fmt.Errorf("failed to upload banner: %w", err)
	}
	return &models.EventBanner{
		Key:         key,
		ContentType: upload.ContentType,
		Size:        int64(len(buf)),
	}, nil
}

// prepareBanner validates the content type and size of an incoming banner,
// buffering the file so its exact size can be recorded and replayed for upload.
func (s *eventService) prepareBanner(upload *models.UploadMedia) ([]byte, error) {
	if !allowedBannerContentTypes[upload.ContentType] {
		return nil, ErrInvalidBannerType
	}
	buf, err := io.ReadAll(io.LimitReader(upload.Reader, maxEventBannerSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read banner: %w", err)
	}
	if len(buf) > maxEventBannerSize {
		return nil, ErrBannerTooLarge
	}
	return buf, nil
}

// generateBannerKey builds a unique, storage-friendly key of the form
// events/{eventID}/banner/{uuid}.{ext}. The original filename is never used
// verbatim as the object key.
func (s *eventService) generateBannerKey(eventID int, filename string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate banner key: %w", err)
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		ext = "bin"
	}
	return fmt.Sprintf("events/%d/banner/%s.%s", eventID, id.String(), ext), nil
}

// cleanupBanner best-effort removes an object that was uploaded but should not
// be referenced anymore (rolled-back writes, replaced or deleted banners).
func (s *eventService) cleanupBanner(ctx context.Context, key string) {
	if key == "" || s.storage == nil {
		return
	}
	if err := s.storage.Delete(ctx, key); err != nil {
		log.Printf("warning: failed to delete event banner %q: %v", key, err)
	}
}

// hydrateBanner resolves an event's stored banner_key into a public URL.
func (s *eventService) hydrateBanner(ctx context.Context, event *models.Event) error {
	if event == nil || event.BannerKey == "" || s.storage == nil {
		return nil
	}
	url, err := s.storage.GetURL(ctx, event.BannerKey)
	if err != nil {
		return err
	}
	event.BannerURL = url
	return nil
}

func (s *eventService) hydrateEvents(ctx context.Context, events []models.Event) error {
	for i := range events {
		if err := s.hydrateBanner(ctx, &events[i]); err != nil {
			return err
		}
	}
	return nil
}
