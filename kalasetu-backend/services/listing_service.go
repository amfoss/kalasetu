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
	ErrNotSeller        = errors.New("only artists and craftspeople can sell listings")
	ErrTitleRequired    = errors.New("title is required")
	ErrInvalidPrice     = errors.New("price must be greater than 0 and at most 9999999999.99")
	ErrInvalidStock     = errors.New("stock must be 0 or more")
	ErrInvalidImages    = errors.New("a listing needs 1 to 8 images, none blank")
	ErrCategoryUnknown  = errors.New("category does not exist")
	ErrListingNotFound  = errors.New("listing not found")
	ErrListingForbidden = errors.New("forbidden: you do not own this listing")
	ErrListingArchived  = errors.New("listing is archived and can no longer be changed")
	ErrInvalidLimit     = errors.New("limit must be at least 1")
	ErrInvalidOffset    = errors.New("offset must be 0 or more")
)

const (
	maxImages = 8
	maxPrice  = 9999999999.99 // NUMERIC(12,2)

	// MaxPageSize caps limit on browse queries; larger requests are clamped.
	MaxPageSize = 100
)

type ListingService interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, userID int, input models.CreateListingInput) (*models.Listing, error)
	// GetByID returns nil, nil when there is no such listing or it is archived
	// and viewerID (0 for anonymous) is not its seller.
	GetByID(ctx context.Context, viewerID, id int) (*models.Listing, error)
	// MyListings returns the seller's own listings, archived ones included.
	MyListings(ctx context.Context, userID int) ([]models.Listing, error)
	// Update applies a partial update with the same validation as Create. Only
	// the seller may update, and not once archived.
	Update(ctx context.Context, userID, id int, input models.UpdateListingInput) (*models.Listing, error)
	// Archive hides the listing from public views; it is one-way and idempotent.
	Archive(ctx context.Context, userID, id int) error
	// Browse returns non-archived listings; a limit above MaxPageSize is clamped.
	Browse(ctx context.Context, q models.ListingQuery) ([]models.Listing, error)
	// Featured returns up to limit random in-stock, non-archived listings.
	Featured(ctx context.Context, limit int) ([]models.Listing, error)
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

func (s *listingService) GetByID(ctx context.Context, viewerID, id int) (*models.Listing, error) {
	l, err := s.listingRepo.FindByID(ctx, id)
	if err != nil || l == nil {
		return nil, err
	}
	if l.Archived && l.Seller.ID != viewerID {
		return nil, nil
	}
	return l, nil
}

func (s *listingService) MyListings(ctx context.Context, userID int) ([]models.Listing, error) {
	return s.listingRepo.FindBySeller(ctx, userID)
}

func (s *listingService) Update(ctx context.Context, userID, id int, input models.UpdateListingInput) (*models.Listing, error) {
	l, err := s.ownedListing(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if l.Archived {
		return nil, ErrListingArchived
	}

	if input.Title != nil {
		t := strings.TrimSpace(*input.Title)
		input.Title = &t
	}
	if input.Description != nil {
		d := strings.TrimSpace(*input.Description)
		input.Description = &d
	}
	if err := validateListing(applyUpdate(l, input)); err != nil {
		return nil, err
	}
	if input.CategoryID != nil {
		if exists, err := s.listingRepo.CategoryExists(ctx, *input.CategoryID); err != nil {
			return nil, err
		} else if !exists {
			return nil, ErrCategoryUnknown
		}
	}

	updated, err := s.listingRepo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, ErrListingArchived // archived since we looked
	}
	return s.listingRepo.FindByID(ctx, id)
}

func (s *listingService) Archive(ctx context.Context, userID, id int) error {
	if _, err := s.ownedListing(ctx, userID, id); err != nil {
		return err
	}
	return s.listingRepo.Archive(ctx, id)
}

// ownedListing loads the listing and checks userID is its seller.
func (s *listingService) ownedListing(ctx context.Context, userID, id int) (*models.Listing, error) {
	l, err := s.listingRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrListingNotFound
	}
	if l.Seller.ID != userID {
		return nil, ErrListingForbidden
	}
	return l, nil
}

// applyUpdate is the listing as it would look after the update, so the update
// can be checked by the same rules as creation.
func applyUpdate(l *models.Listing, in models.UpdateListingInput) models.CreateListingInput {
	merged := models.CreateListingInput{
		Title: l.Title, Description: l.Description, Price: l.Price,
		Stock: l.Stock, ImageURLs: l.ImageURLs, CategoryID: l.Category.ID,
	}
	if in.Title != nil {
		merged.Title = *in.Title
	}
	if in.Description != nil {
		merged.Description = *in.Description
	}
	if in.Price != nil {
		merged.Price = *in.Price
	}
	if in.Stock != nil {
		merged.Stock = *in.Stock
	}
	if in.ImageURLs != nil {
		merged.ImageURLs = in.ImageURLs
	}
	if in.CategoryID != nil {
		merged.CategoryID = *in.CategoryID
	}
	return merged
}

func (s *listingService) Browse(ctx context.Context, q models.ListingQuery) ([]models.Listing, error) {
	if q.Limit < 1 {
		return nil, ErrInvalidLimit
	}
	if q.Offset < 0 {
		return nil, ErrInvalidOffset
	}
	q.Limit = min(q.Limit, MaxPageSize)
	q.Query = strings.TrimSpace(q.Query)
	return s.listingRepo.Search(ctx, q)
}

func (s *listingService) Featured(ctx context.Context, limit int) ([]models.Listing, error) {
	if limit < 1 {
		return nil, ErrInvalidLimit
	}
	return s.listingRepo.Featured(ctx, min(limit, MaxPageSize))
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
