package services

import (
	"context"
	"fmt"
	"time"

	"kalasetu/models"
	"kalasetu/payments"
	"kalasetu/repos"
)

type CheckoutSessionService interface {
	// Create validates the Buyer's Cart and Shipping address, reserves Stock
	// and opens a payment order with the gateway, returning everything the
	// Buyer's browser needs to open the payment screen. See
	// repos.CheckoutSessionRepository.Create for the failure modes.
	Create(ctx context.Context, buyerID int, ship models.ShippingAddress) (*models.CheckoutSession, error)
}

type checkoutSessionService struct {
	repo              repos.CheckoutSessionRepository
	gateway           payments.Gateway
	keyID             string
	reservationWindow time.Duration
}

// NewCheckoutSessionService builds a CheckoutSessionService. keyID is the
// gateway's public key, handed back to the Buyer's browser as-is; it is not
// a secret.
func NewCheckoutSessionService(repo repos.CheckoutSessionRepository, gateway payments.Gateway, keyID string, reservationWindow time.Duration) CheckoutSessionService {
	return &checkoutSessionService{repo: repo, gateway: gateway, keyID: keyID, reservationWindow: reservationWindow}
}

func (s *checkoutSessionService) Create(ctx context.Context, buyerID int, ship models.ShippingAddress) (*models.CheckoutSession, error) {
	ship, err := normalizeShipping(ship)
	if err != nil {
		return nil, err
	}

	createOrder := func(ctx context.Context, amountCents int64, reference string) (string, error) {
		res, err := s.gateway.CreatePaymentOrder(ctx, payments.CreateOrderRequest{
			Amount:   amountCents,
			Currency: "INR",
			Receipt:  reference,
			Capture:  payments.CaptureAutomatic,
		})
		if err != nil {
			return "", fmt.Errorf("could not open a payment: %w", err)
		}
		return res.GatewayOrderID, nil
	}

	session, err := s.repo.Create(ctx, buyerID, ship, time.Now().Add(s.reservationWindow), createOrder)
	if err != nil {
		return nil, err
	}
	session.KeyID = s.keyID
	return session, nil
}
