package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"kalasetu/models"
	"kalasetu/payments"
	"kalasetu/repos"
)

var (
	ErrShippingAddressIncomplete = errors.New("shipping address is incomplete")
	ErrOrderItemNotFound         = errors.New("order item not found")
	ErrOrderItemForbidden        = errors.New("forbidden: this order item is not for one of your listings")
	ErrInvalidStatusTransition   = errors.New("invalid status transition: an order item goes paid, shipped, delivered, one step at a time")
)

type OrderService interface {
	// Checkout turns the buyer's Cart into an Order, all or nothing. See
	// repos.OrderRepository.Checkout for the failure modes.
	Checkout(ctx context.Context, buyerID int, ship models.ShippingAddress) (*models.Order, error)
	// MyOrders returns the buyer's Orders, newest first.
	MyOrders(ctx context.Context, buyerID int) ([]models.Order, error)
	// SellerOrderItems returns the Order Items for the seller's Listings, newest
	// first, optionally only those in status.
	SellerOrderItems(ctx context.Context, sellerID int, status *models.FulfilmentStatus) ([]models.SellerOrderItem, error)
	// UpdateItemStatus lets the item's Seller move it forward one step (paid to
	// shipped to delivered). Anyone else gets ErrOrderItemForbidden; skipping,
	// repeating or reversing gets ErrInvalidStatusTransition.
	UpdateItemStatus(ctx context.Context, sellerID, itemID int, status models.FulfilmentStatus) (*models.SellerOrderItem, error)
}

type orderService struct {
	repo     repos.OrderRepository
	payments payments.PaymentProvider
}

func NewOrderService(repo repos.OrderRepository, provider payments.PaymentProvider) OrderService {
	return &orderService{repo: repo, payments: provider}
}

func (s *orderService) Checkout(ctx context.Context, buyerID int, ship models.ShippingAddress) (*models.Order, error) {
	ship.Name, ship.Phone = strings.TrimSpace(ship.Name), strings.TrimSpace(ship.Phone)
	ship.Line1, ship.Line2 = strings.TrimSpace(ship.Line1), strings.TrimSpace(ship.Line2)
	ship.City, ship.State = strings.TrimSpace(ship.City), strings.TrimSpace(ship.State)
	ship.PostalCode, ship.Country = strings.TrimSpace(ship.PostalCode), strings.TrimSpace(ship.Country)
	// line2 is the only optional field.
	for _, v := range []string{ship.Name, ship.Phone, ship.Line1, ship.City, ship.State, ship.PostalCode, ship.Country} {
		if v == "" {
			return nil, ErrShippingAddressIncomplete
		}
	}

	var charged *payments.RefundRequest
	pay := func(ctx context.Context, amountCents int64, reference string) (string, error) {
		res, err := s.payments.Charge(ctx, payments.ChargeRequest{Amount: amountCents, Reference: reference})
		if err != nil {
			return "", fmt.Errorf("payment failed: %w", err)
		}
		charged = &payments.RefundRequest{ChargeID: res.ChargeID, Amount: amountCents}
		return res.ChargeID, nil
	}
	order, err := s.repo.Checkout(ctx, buyerID, ship, pay)

	if err != nil && charged != nil {
		// Charged but the transaction did not commit: give the money back.
		_ = s.payments.Refund(ctx, *charged)
	}
	return order, err
}

func (s *orderService) MyOrders(ctx context.Context, buyerID int) ([]models.Order, error) {
	return s.repo.FindByBuyer(ctx, buyerID)
}

func (s *orderService) SellerOrderItems(ctx context.Context, sellerID int, status *models.FulfilmentStatus) ([]models.SellerOrderItem, error) {
	return s.repo.FindItemsBySeller(ctx, sellerID, status)
}

func (s *orderService) UpdateItemStatus(ctx context.Context, sellerID, itemID int, status models.FulfilmentStatus) (*models.SellerOrderItem, error) {
	item, err := s.repo.FindItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrOrderItemNotFound
	}
	if item.SellerID != sellerID {
		return nil, ErrOrderItemForbidden
	}
	if !item.Status.CanAdvanceTo(status) {
		return nil, ErrInvalidStatusTransition
	}
	ok, err := s.repo.AdvanceItem(ctx, itemID, item.Status, status)
	if err != nil {
		return nil, err
	}
	if !ok {
		// Another request moved the item after we read it.
		return nil, ErrInvalidStatusTransition
	}
	item.Status = status
	return item, nil
}
