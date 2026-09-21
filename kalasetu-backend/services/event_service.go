package services

import (
	"context"
	"errors"
	"kalasetu/models"
	"kalasetu/repos"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrForbidden     = errors.New("you are not the host of this event")
)

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
}

func NewEventService(eventRepo repos.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
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
	// Refetch so the response includes host_name and DB-normalized fields.
	return s.eventRepo.FindByID(ctx, event.ID)
}

func (s *eventService) List(ctx context.Context) ([]models.Event, error) {
	return s.eventRepo.List(ctx)
}
func (s *eventService) ListByUser(ctx context.Context, userID int) ([]models.Event, error) {
	return s.eventRepo.ListByUser(ctx, userID)
}

func (s *eventService) GetByID(ctx context.Context, id int) (*models.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
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

	if err := s.eventRepo.Update(ctx, id, input); err != nil {
		return nil, err
	}
	return s.eventRepo.FindByID(ctx, id)
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
	return s.eventRepo.Delete(ctx, id)
}