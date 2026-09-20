package services

import (
	"context"
	"errors"
	"math"

	"kalasetu/models"
	"kalasetu/repos"
)

var (
	ErrCartOwnListing     = errors.New("you cannot add your own listing to your cart")
	ErrCartListingArchive = errors.New("listing is archived and cannot be added to a cart")
	ErrCartQuantityAdd    = errors.New("quantity must be at least 1")
	ErrCartQuantityUpdate = errors.New("quantity must be 0 or more")
	ErrCartOverStock      = errors.New("quantity exceeds the available stock")
	ErrCartItemNotFound   = errors.New("listing is not in your cart")
)

type CartService interface {
	// Get returns the user's cart with current Listing details. Lines whose
	// Listing became archived or short of stock are kept and flagged.
	Get(ctx context.Context, userID int) (*models.Cart, error)
	// Add increases the quantity of the listing's line by quantity. The result
	// may not exceed the listing's stock, and the listing must be live and not
	// the user's own.
	Add(ctx context.Context, userID, listingID, quantity int) (*models.Cart, error)
	// SetQuantity sets an existing line's quantity; 0 removes the line.
	SetQuantity(ctx context.Context, userID, listingID, quantity int) (*models.Cart, error)
	// Remove drops the listing's line; an absent line is a no-op.
	Remove(ctx context.Context, userID, listingID int) (*models.Cart, error)
}

type cartService struct {
	cartRepo    repos.CartRepository
	listingRepo repos.ListingRepository
}

func NewCartService(cartRepo repos.CartRepository, listingRepo repos.ListingRepository) CartService {
	return &cartService{cartRepo: cartRepo, listingRepo: listingRepo}
}

func (s *cartService) Get(ctx context.Context, userID int) (*models.Cart, error) {
	items, err := s.cartRepo.Items(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(items))
	for i, it := range items {
		ids[i] = it.ListingID
	}
	listings, err := s.listingRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]models.Listing, len(listings))
	for _, l := range listings {
		byID[l.ID] = l
	}

	cart := &models.Cart{Lines: make([]models.CartLine, 0, len(items))}
	var cents int64
	for _, it := range items {
		l, ok := byID[it.ListingID]
		if !ok {
			continue // deleted meanwhile; the foreign key removes the line
		}
		line := models.CartLine{Listing: l, Quantity: it.Quantity, Issue: lineIssue(l, it.Quantity)}
		cart.Lines = append(cart.Lines, line)
		if line.Issue == "" {
			cents += int64(math.Round(l.Price*100)) * int64(it.Quantity)
		}
	}
	cart.Total = float64(cents) / 100
	return cart, nil
}

func lineIssue(l models.Listing, quantity int) models.CartLineIssue {
	switch {
	case l.Archived:
		return models.CartIssueArchived
	case l.Stock == 0:
		return models.CartIssueOutOfStock
	case l.Stock < quantity:
		return models.CartIssueInsufficientStock
	}
	return ""
}

func (s *cartService) Add(ctx context.Context, userID, listingID, quantity int) (*models.Cart, error) {
	if quantity < 1 {
		return nil, ErrCartQuantityAdd
	}
	l, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrListingNotFound
	}
	if l.Seller.ID == userID {
		return nil, ErrCartOwnListing
	}
	if l.Archived {
		return nil, ErrCartListingArchive
	}
	current, err := s.cartRepo.Quantity(ctx, userID, listingID)
	if err != nil {
		return nil, err
	}
	if current+quantity > l.Stock {
		return nil, ErrCartOverStock
	}
	if err := s.cartRepo.Set(ctx, userID, listingID, current+quantity); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}

func (s *cartService) SetQuantity(ctx context.Context, userID, listingID, quantity int) (*models.Cart, error) {
	if quantity < 0 {
		return nil, ErrCartQuantityUpdate
	}
	if quantity == 0 {
		return s.Remove(ctx, userID, listingID)
	}
	current, err := s.cartRepo.Quantity(ctx, userID, listingID)
	if err != nil {
		return nil, err
	}
	if current == 0 {
		return nil, ErrCartItemNotFound
	}
	l, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrListingNotFound
	}
	if l.Archived {
		return nil, ErrCartListingArchive
	}
	if quantity > l.Stock {
		return nil, ErrCartOverStock
	}
	if err := s.cartRepo.Set(ctx, userID, listingID, quantity); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}

func (s *cartService) Remove(ctx context.Context, userID, listingID int) (*models.Cart, error) {
	if err := s.cartRepo.Remove(ctx, userID, listingID); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}
