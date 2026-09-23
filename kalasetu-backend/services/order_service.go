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
	ErrOrderItemCancelForbidden  = errors.New("forbidden: only the buyer or the seller can cancel this order item")
	ErrRefundFailed              = errors.New("refund failed")
	ErrInvalidStatusTransition   = errors.New("invalid status transition: an order item goes paid, shipped, delivered, one step at a time")
	// ErrPaymentConfirmationInvalid means the gateway could not verify the
	// confirmation's signature: it is tampered, forged, or does not match a
	// payment the gateway actually processed.
	ErrPaymentConfirmationInvalid = errors.New("payment confirmation could not be verified")
)

// refundWebhookStatus maps a Razorpay refund webhook's event type to the
// RefundStatus it advances the matching Order Item to. The handler only
// calls AdvanceRefundFromWebhook for event types it knows are in here.
var refundWebhookStatus = map[string]models.RefundStatus{
	"refund.processed": models.RefundSettled,
	"refund.failed":    models.RefundFailed,
}

type OrderService interface {
	// ConfirmCheckoutSessionPayment verifies confirmation against the gateway
	// before anything is created, then fulfils the caller's Checkout Session
	// into an Order. See repos.OrderRepository.ConfirmCheckoutSession for the
	// failure modes; a failed signature verification returns
	// ErrPaymentConfirmationInvalid. A late payment whose Stock could not be
	// re-acquired is refunded automatically and reported as
	// repos.ErrCheckoutSessionRefunded rather than an Order.
	ConfirmCheckoutSessionPayment(ctx context.Context, buyerID int, confirmation payments.ConfirmationRequest) (*models.Order, error)
	// FulfilFromWebhook is the webhook backstop's entry into the same
	// fulfilment ConfirmCheckoutSessionPayment runs, for a payment-captured
	// or order-paid notification whose signature the caller has already
	// verified. amountCents is the gateway's own account of the payment,
	// used only if gatewayOrderID matches no Checkout Session at all. See
	// repos.OrderRepository.FulfilCheckoutSessionFromWebhook for the
	// delivered/idempotency contract and how a late payment or an unmatched
	// payment is refunded automatically rather than surfaced as an error.
	FulfilFromWebhook(ctx context.Context, eventID, eventType, gatewayOrderID, paymentID string, amountCents int64) (order *models.Order, delivered bool, err error)
	// AdvanceRefundFromWebhook is a refund-processed or refund-failed
	// notification's entry: it advances the Order Item whose refund_id
	// matches refundID to the RefundStatus eventType denotes. See
	// repos.OrderRepository.AdvanceRefundFromWebhook for the
	// delivered/idempotency contract and why an unmatched refundID is not an
	// error.
	AdvanceRefundFromWebhook(ctx context.Context, eventID, eventType, refundID string) (delivered bool, err error)
	// ReleaseFromWebhook is a payment-failed notification's entry: it
	// releases the Checkout Session matching gatewayOrderID immediately
	// rather than waiting for it to expire. See
	// repos.OrderRepository.ReleaseCheckoutSessionFromWebhook for the
	// delivered/idempotency contract.
	ReleaseFromWebhook(ctx context.Context, eventID, eventType, gatewayOrderID string) (delivered bool, err error)
	// MyOrders returns the buyer's Orders, newest first.
	MyOrders(ctx context.Context, buyerID int) ([]models.Order, error)
	// SellerOrderItems returns the Order Items for the seller's Listings, newest
	// first, optionally only those in status.
	SellerOrderItems(ctx context.Context, sellerID int, status *models.FulfilmentStatus) ([]models.SellerOrderItem, error)
	// UpdateItemStatus lets the item's Seller move it forward one step (paid to
	// shipped to delivered). Anyone else gets ErrOrderItemForbidden; skipping,
	// repeating or reversing gets ErrInvalidStatusTransition.
	UpdateItemStatus(ctx context.Context, sellerID, itemID int, status models.FulfilmentStatus) (*models.SellerOrderItem, error)
	// CancelItem lets the item's Buyer or Seller cancel it while it is paid,
	// restocking the Listing and refunding through the payment gateway. Anyone
	// else gets ErrOrderItemCancelForbidden, a non-paid item
	// repos.ErrItemNotCancellable, and a failed refund ErrRefundFailed with
	// nothing changed. Refunds carry a stable idempotency key derived from the
	// item, so retrying a cancel that failed after its refund does not refund
	// twice; a refund the gateway has accepted but not yet settled, or a
	// conflict meaning it was already accepted, both count as success.
	CancelItem(ctx context.Context, userID, itemID int) (*models.SellerOrderItem, error)
}

type orderService struct {
	repo    repos.OrderRepository
	gateway payments.Gateway
}

// normalizeShipping trims ship's fields and rejects it if any but the
// optional line2 is blank.
func normalizeShipping(ship models.ShippingAddress) (models.ShippingAddress, error) {
	ship.Name, ship.Phone = strings.TrimSpace(ship.Name), strings.TrimSpace(ship.Phone)
	ship.Line1, ship.Line2 = strings.TrimSpace(ship.Line1), strings.TrimSpace(ship.Line2)
	ship.City, ship.State = strings.TrimSpace(ship.City), strings.TrimSpace(ship.State)
	ship.PostalCode, ship.Country = strings.TrimSpace(ship.PostalCode), strings.TrimSpace(ship.Country)
	// line2 is the only optional field.
	for _, v := range []string{ship.Name, ship.Phone, ship.Line1, ship.City, ship.State, ship.PostalCode, ship.Country} {
		if v == "" {
			return models.ShippingAddress{}, ErrShippingAddressIncomplete
		}
	}
	return ship, nil
}

func NewOrderService(repo repos.OrderRepository, gateway payments.Gateway) OrderService {
	return &orderService{repo: repo, gateway: gateway}
}

func (s *orderService) ConfirmCheckoutSessionPayment(ctx context.Context, buyerID int, confirmation payments.ConfirmationRequest) (*models.Order, error) {
	if err := s.gateway.VerifyConfirmation(ctx, confirmation); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentConfirmationInvalid, err)
	}
	return s.repo.ConfirmCheckoutSession(ctx, buyerID, confirmation.GatewayOrderID, confirmation.PaymentID, s.refundFunc())
}

func (s *orderService) FulfilFromWebhook(ctx context.Context, eventID, eventType, gatewayOrderID, paymentID string, amountCents int64) (*models.Order, bool, error) {
	return s.repo.FulfilCheckoutSessionFromWebhook(ctx, eventID, eventType, gatewayOrderID, paymentID, amountCents, s.refundFunc())
}

// refundFunc adapts the gateway's Refund call into a repos.RefundFunc, the
// shape every refunding repo method (cancellation, a late payment, an
// unmatched payment) takes so the actual gateway call and its error mapping
// live in exactly one place.
func (s *orderService) refundFunc() repos.RefundFunc {
	return func(ctx context.Context, chargeID string, amountCents int64, idempotencyKey string) (repos.RefundResult, error) {
		res, err := s.gateway.Refund(ctx, payments.RefundOrderRequest{
			PaymentID:      chargeID,
			Amount:         amountCents,
			IdempotencyKey: idempotencyKey,
			Reference:      idempotencyKey,
		})
		if err != nil {
			return repos.RefundResult{}, fmt.Errorf("%w: %w", ErrRefundFailed, err)
		}
		return repos.RefundResult{RefundID: res.RefundID, Status: res.Status}, nil
	}
}

func (s *orderService) ReleaseFromWebhook(ctx context.Context, eventID, eventType, gatewayOrderID string) (bool, error) {
	return s.repo.ReleaseCheckoutSessionFromWebhook(ctx, eventID, eventType, gatewayOrderID)
}

func (s *orderService) AdvanceRefundFromWebhook(ctx context.Context, eventID, eventType, refundID string) (bool, error) {
	status, ok := refundWebhookStatus[eventType]
	if !ok {
		return false, fmt.Errorf("order service: unhandled refund webhook event type %q", eventType)
	}
	return s.repo.AdvanceRefundFromWebhook(ctx, eventID, eventType, refundID, status)
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

func (s *orderService) CancelItem(ctx context.Context, userID, itemID int) (*models.SellerOrderItem, error) {
	item, err := s.repo.FindItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrOrderItemNotFound
	}
	if userID != item.BuyerID && userID != item.SellerID {
		return nil, ErrOrderItemCancelForbidden
	}
	result, err := s.repo.CancelItem(ctx, itemID, s.refundFunc())
	if err != nil {
		return nil, err
	}
	item.Status = models.StatusCancelled
	item.RefundID = result.RefundID
	item.RefundStatus = models.RefundAccepted
	return item, nil
}
