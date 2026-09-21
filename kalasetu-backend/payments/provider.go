package payments

import (
	"context"
	"errors"
)

// PaymentProvider is the seam between the marketplace and a payment gateway.
// Amounts are in the smallest currency unit (e.g. paise/cents).
type PaymentProvider interface {
	Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
	Refund(ctx context.Context, req RefundRequest) error
}

type ChargeRequest struct {
	Amount    int64
	Reference string // caller's identifier for what is being paid for, e.g. an order id
}

type ChargeResult struct {
	ChargeID string
}

type RefundRequest struct {
	ChargeID  string
	Amount    int64
	Reference string // caller's identifier for this refund, e.g. an order item id
}

// ErrNotConfigured is returned by the provider used when no real gateway is wired in.
var ErrNotConfigured = errors.New("payment provider not configured")

type unconfigured struct{}

// NewUnconfigured returns a provider whose every call fails with ErrNotConfigured.
// It is the production placeholder until a real gateway is integrated.
func NewUnconfigured() PaymentProvider { return unconfigured{} }

func (unconfigured) Charge(context.Context, ChargeRequest) (ChargeResult, error) {
	return ChargeResult{}, ErrNotConfigured
}

func (unconfigured) Refund(context.Context, RefundRequest) error { return ErrNotConfigured }
