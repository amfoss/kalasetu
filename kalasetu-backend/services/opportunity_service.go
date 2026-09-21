package services

import (
	"context"
	"errors"
	"strings"

	"kalasetu/models"
	"kalasetu/repos"
)

var (
	ErrOpportunityNotFound = errors.New("opportunity not found")
	ErrTitleRequired       = errors.New("title is required")
)

type OpportunityService interface {
	Create(ctx context.Context, userID int, input models.CreateOpportunityInput) (*models.Opportunity, error)
	Update(ctx context.Context, userID, id int, input models.UpdateOpportunityInput) (*models.Opportunity, error)
	FindByID(ctx context.Context, id int) (*models.Opportunity, error)
	ListByEvent(ctx context.Context, eventID int) ([]models.Opportunity, error)
}

type opportunityService struct {
	opportunityRepo repos.OpportunityRepository
}

func NewOpportunityService(opportunityRepo repos.OpportunityRepository) OpportunityService {
	return &opportunityService{
		opportunityRepo: opportunityRepo,
	}
}

func (s *opportunityService) Create(
	ctx context.Context,
	userID int,
	input models.CreateOpportunityInput,
) (*models.Opportunity, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, ErrTitleRequired
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if status == "" {
		status = "ACTIVE"
	}

	totalPositions := input.TotalPositions
	if totalPositions <= 0 {
		totalPositions = 1
	}

	opp := &models.Opportunity{
		EventID:        input.EventID,
		Title:          strings.TrimSpace(input.Title),
		Description:    strings.TrimSpace(input.Description),
		Categories:     strings.Join(input.Categories, ","),
		Location:       strings.TrimSpace(input.Location),
		StartDate:      strings.TrimSpace(input.StartDate),
		EndDate:        strings.TrimSpace(input.EndDate),
		TotalPositions: totalPositions,
		Status:         status,
		HostID:         userID,
	}

	err := s.opportunityRepo.Create(ctx, opp)
	if err != nil {
		return nil, err
	}

	return opp, nil
}

func (s *opportunityService) FindByID(ctx context.Context, id int) (*models.Opportunity, error) {
	opp, err := s.opportunityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if opp == nil {
		return nil, ErrOpportunityNotFound
	}
	return opp, nil
}

func (s *opportunityService) ListByEvent(ctx context.Context, eventID int) ([]models.Opportunity, error) {
	return s.opportunityRepo.ListByEvent(ctx, eventID)
}

func (s *opportunityService) Update(
	ctx context.Context,
	userID, id int,
	input models.UpdateOpportunityInput,
) (*models.Opportunity, error) {
	opp, err := s.opportunityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if opp == nil {
		return nil, ErrOpportunityNotFound
	}

	if input.Title != nil && strings.TrimSpace(*input.Title) != "" {
		opp.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		opp.Description = strings.TrimSpace(*input.Description)
	}
	if len(input.Categories) > 0 {
		opp.Categories = strings.Join(input.Categories, ",")
	}
	if input.Location != nil {
		opp.Location = strings.TrimSpace(*input.Location)
	}
	if input.StartDate != nil {
		opp.StartDate = strings.TrimSpace(*input.StartDate)
	}
	if input.TotalPositions != nil && *input.TotalPositions > 0 {
		opp.TotalPositions = *input.TotalPositions
	}
	if input.Status != nil && strings.TrimSpace(*input.Status) != "" {
		opp.Status = strings.ToUpper(strings.TrimSpace(*input.Status))
	}

	if err := s.opportunityRepo.Update(ctx, opp); err != nil {
		return nil, err
	}

	return opp, nil
}

