package payments

import (
	"context"
	"encoding/json"
)

// Gateway is the shape KalaSetu needs from a payment gateway for a two-phase
// flow: create an order at the gateway, verify a confirmation the Buyer
// relays back after paying, refund a settled payment, and parse a signed
// webhook the gateway sends out of band. It is not a claim that gateways are
// interchangeable behind it - it documents what KalaSetu itself requires.
type Gateway interface {
	// CreatePaymentOrder registers an order with the gateway before the
	// Buyer pays, and returns the gateway's identifier for it.
	CreatePaymentOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResult, error)

	// VerifyConfirmation checks a payment confirmation the Buyer relayed
	// back from the gateway's checkout flow. A nil error means the
	// confirmation is authentic.
	VerifyConfirmation(ctx context.Context, req ConfirmationRequest) error

	// Refund requests a refund of a previously captured payment.
	Refund(ctx context.Context, req RefundOrderRequest) (RefundOrderResult, error)

	// ParseWebhook verifies and decodes a webhook payload the gateway sent.
	ParseWebhook(ctx context.Context, rawBody []byte, signature string) (WebhookEvent, error)
}

// CaptureMode controls whether a payment order is captured automatically
// once the Buyer pays, or held for a manual capture later.
type CaptureMode int

const (
	CaptureAutomatic CaptureMode = iota
	CaptureManual
)

type CreateOrderRequest struct {
	// Amount is in the smallest currency unit (e.g. paise/cents).
	Amount   int64
	Currency string
	// Receipt is the caller's identifier for what is being paid for, e.g.
	// an order id.
	Receipt string
	Capture CaptureMode
}

type CreateOrderResult struct {
	// GatewayOrderID is the gateway's identifier for the order, to be
	// handed to the Buyer's client to start the gateway's checkout.
	GatewayOrderID string
}

// ConfirmationRequest is what the Buyer's client relays back once they pay:
// the gateway's order and payment identifiers, and a signature proving the
// confirmation came from the gateway rather than a tampering client.
type ConfirmationRequest struct {
	GatewayOrderID string
	PaymentID      string
	Signature      string
}

type RefundOrderRequest struct {
	PaymentID string
	// Amount is in the smallest currency unit (e.g. paise/cents).
	Amount int64
	// IdempotencyKey deduplicates retried refund requests at the gateway.
	// It must be at least ten characters drawn from alphanumerics,
	// hyphens and underscores.
	IdempotencyKey string
	// Reference is the caller's identifier for this refund, e.g. an order
	// item id.
	Reference string
}

type RefundOrderResult struct {
	RefundID string
	// Status distinguishes accepted-but-pending refunds from ones the
	// gateway has already settled.
	Status string
}

// WebhookEvent is a gateway webhook payload once its signature has been
// verified.
type WebhookEvent struct {
	Event   string
	Payload json.RawMessage
}

type unconfiguredGateway struct{}

// NewUnconfiguredGateway returns a Gateway whose every call fails with
// ErrNotConfigured. It is the production placeholder until a real gateway is
// wired in.
func NewUnconfiguredGateway() Gateway { return unconfiguredGateway{} }

func (unconfiguredGateway) CreatePaymentOrder(context.Context, CreateOrderRequest) (CreateOrderResult, error) {
	return CreateOrderResult{}, ErrNotConfigured
}

func (unconfiguredGateway) VerifyConfirmation(context.Context, ConfirmationRequest) error {
	return ErrNotConfigured
}

func (unconfiguredGateway) Refund(context.Context, RefundOrderRequest) (RefundOrderResult, error) {
	return RefundOrderResult{}, ErrNotConfigured
}

func (unconfiguredGateway) ParseWebhook(context.Context, []byte, string) (WebhookEvent, error) {
	return WebhookEvent{}, ErrNotConfigured
}
