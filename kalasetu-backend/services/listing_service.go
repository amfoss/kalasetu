package services

import (
	"context"
	"errors"
	"math"
	"strings"

	"kalasetu/models"
	"kalasetu/repos"
)

var (
	ErrNotSeller       = errors.New("only artists and craftspeople can sell listings")
	ErrTitleRequired   = errors.New("title is required")
	ErrInvalidPrice    = errors.New("price must be greater than 0 and at most 9999999999.99")
	ErrInvalidStock    = errors.New("stock must be 0 or more")
	ErrInvalidImages   = errors.New("a listing needs 1 to 8 images, none blank")
	ErrCategoryUnknown = errors.New("category does not exist")
)

const (
	maxImages = 8
	maxPrice  = 9999999999.99 // NUMERIC(12,2)
)

type ListingService interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, userID int, input models.CreateListingInput) (*models.Listing, error)
	// GetByID returns nil, nil when there is no such listing.
	GetByID(ctx context.Context, id int) (*models.Listing, error)
}

type listingService struct {
	listingRepo repos.ListingRepository
}

func NewListingService(listingRepo repos.ListingRepository) ListingService {
	return &listingService{listingRepo: listingRepo}
}

func (s *listingService) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.listingRepo.ListCategories(ctx)
}

func (s *listingService) Create(ctx context.Context, userID int, input models.CreateListingInput) (*models.Listing, error) {
	isSeller, err := s.listingRepo.UserHasAnyRole(ctx, userID, "Artist", "Craftsperson")
	if err != nil {
		return nil, err
	}
	if !isSeller {
		return nil, ErrNotSeller
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if err := validateListing(input); err != nil {
		return nil, err
	}
	if exists, err := s.listingRepo.CategoryExists(ctx, input.CategoryID); err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrCategoryUnknown
	}

	id, err := s.listingRepo.Create(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	return s.listingRepo.FindByID(ctx, id)
}

func (s *listingService) GetByID(ctx context.Context, id int) (*models.Listing, error) {
	return s.listingRepo.FindByID(ctx, id)
}

func validateListing(in models.CreateListingInput) error {
	if in.Title == "" {
		return ErrTitleRequired
	}
	// Compare the rounded value so 0.004 (stored as 0.00) is rejected too.
	if rounded := math.Round(in.Price*100) / 100; !(rounded > 0 && rounded <= maxPrice) {
		return ErrInvalidPrice
	}
	if in.Stock < 0 {
		return ErrInvalidStock
	}
	if len(in.ImageURLs) < 1 || len(in.ImageURLs) > maxImages {
		return ErrInvalidImages
	}
	for _, u := range in.ImageURLs {
		if strings.TrimSpace(u) == "" {
			return ErrInvalidImages
		}
	}
	return nil
}
