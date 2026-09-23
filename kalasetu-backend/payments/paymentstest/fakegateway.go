package paymentstest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"kalasetu/payments"
)

// fakeGatewaySecret keys FakeGateway's HMAC signatures. It plays the role a
// real gateway's API secret would, entirely in memory.
const fakeGatewaySecret = "fake-gateway-secret"

// errFakeInvalidSignature is returned by VerifyConfirmation and ParseWebhook
// when the given signature does not match what FakeGateway itself minted.
var errFakeInvalidSignature = errors.New("paymentstest: invalid confirmation signature")

// FakeGateway is an in-memory payments.Gateway for tests. It succeeds unless
// told to fail, and can mint a confirmation that passes its own
// VerifyConfirmation, so a test can express "the Buyer paid" without any
// HTTP.
type FakeGateway struct {
	mu sync.Mutex

	createOrderErr   error
	verifyErr        error
	refundErr        error
	nextRefundStatus string

	nextOrderID   int
	nextPaymentID int
	nextRefundID  int

	orders   []payments.CreateOrderRequest
	refunds  []payments.RefundOrderRequest
	refunded map[string]payments.RefundOrderResult // idempotency key -> result
}

func NewFakeGateway() *FakeGateway {
	return &FakeGateway{refunded: map[string]payments.RefundOrderResult{}}
}

var _ payments.Gateway = (*FakeGateway)(nil)

// FailCreateOrder makes subsequent CreatePaymentOrder calls return err; nil restores success.
func (f *FakeGateway) FailCreateOrder(err error) { f.mu.Lock(); f.createOrderErr = err; f.mu.Unlock() }

// FailVerifyConfirmation makes subsequent VerifyConfirmation calls return err; nil restores success.
func (f *FakeGateway) FailVerifyConfirmation(err error) {
	f.mu.Lock()
	f.verifyErr = err
	f.mu.Unlock()
}

// FailRefund makes subsequent Refund calls return err; nil restores success.
func (f *FakeGateway) FailRefund(err error) { f.mu.Lock(); f.refundErr = err; f.mu.Unlock() }

// NextRefundStatus makes the next fresh (non-replayed) Refund call return
// status instead of "processed", the way a real gateway reports a refund it
// has accepted but not yet settled. It applies once.
func (f *FakeGateway) NextRefundStatus(status string) {
	f.mu.Lock()
	f.nextRefundStatus = status
	f.mu.Unlock()
}

// SeedRefund records result as the outcome already on file for idempotencyKey,
// so the next Refund call using it replays result instead of creating a new
// refund - standing in for a real gateway's conflict response to a repeated
// idempotency key.
func (f *FakeGateway) SeedRefund(idempotencyKey string, result payments.RefundOrderResult) {
	f.mu.Lock()
	f.refunded[idempotencyKey] = result
	f.mu.Unlock()
}

// Orders returns the orders created so far.
func (f *FakeGateway) Orders() []payments.CreateOrderRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]payments.CreateOrderRequest(nil), f.orders...)
}

// Refunds returns the successful refunds so far.
func (f *FakeGateway) Refunds() []payments.RefundOrderRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]payments.RefundOrderRequest(nil), f.refunds...)
}

func (f *FakeGateway) CreatePaymentOrder(_ context.Context, req payments.CreateOrderRequest) (payments.CreateOrderResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createOrderErr != nil {
		return payments.CreateOrderResult{}, f.createOrderErr
	}
	f.nextOrderID++
	f.orders = append(f.orders, req)
	return payments.CreateOrderResult{GatewayOrderID: fmt.Sprintf("fake_order_%d", f.nextOrderID)}, nil
}

// Confirm mints a payment confirmation for gatewayOrderID that will pass
// VerifyConfirmation, standing in for the Buyer actually paying at the
// gateway's checkout.
func (f *FakeGateway) Confirm(gatewayOrderID string) payments.ConfirmationRequest {
	f.mu.Lock()
	f.nextPaymentID++
	paymentID := fmt.Sprintf("fake_payment_%d", f.nextPaymentID)
	f.mu.Unlock()

	return payments.ConfirmationRequest{
		GatewayOrderID: gatewayOrderID,
		PaymentID:      paymentID,
		Signature:      fakeSignature(gatewayOrderID + "|" + paymentID),
	}
}

func (f *FakeGateway) VerifyConfirmation(_ context.Context, req payments.ConfirmationRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.verifyErr != nil {
		return f.verifyErr
	}
	expected := fakeSignature(req.GatewayOrderID + "|" + req.PaymentID)
	if !hmac.Equal([]byte(req.Signature), []byte(expected)) {
		return errFakeInvalidSignature
	}
	return nil
}

// Refund records the refund, or replays the recorded result without
// recording a second one when req.IdempotencyKey has already been used, the
// way a real gateway treats a repeated idempotency key.
func (f *FakeGateway) Refund(_ context.Context, req payments.RefundOrderRequest) (payments.RefundOrderResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.refundErr != nil {
		return payments.RefundOrderResult{}, f.refundErr
	}
	if req.IdempotencyKey != "" {
		if result, ok := f.refunded[req.IdempotencyKey]; ok {
			return result, nil
		}
	}
	f.nextRefundID++
	status := "processed"
	if f.nextRefundStatus != "" {
		status = f.nextRefundStatus
		f.nextRefundStatus = ""
	}
	result := payments.RefundOrderResult{
		RefundID: fmt.Sprintf("fake_refund_%d", f.nextRefundID),
		Status:   status,
	}
	f.refunds = append(f.refunds, req)
	if req.IdempotencyKey != "" {
		f.refunded[req.IdempotencyKey] = result
	}
	return result, nil
}

// ParseWebhook verifies a payload minted by SignWebhook and decodes it,
// mirroring a real gateway's webhook contract without any HTTP.
func (f *FakeGateway) ParseWebhook(_ context.Context, rawBody []byte, signature string) (payments.WebhookEvent, error) {
	if !hmac.Equal([]byte(signature), []byte(fakeSignature(string(rawBody)))) {
		return payments.WebhookEvent{}, errFakeInvalidSignature
	}
	var decoded struct {
		Event   string          `json:"event"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(rawBody, &decoded); err != nil {
		return payments.WebhookEvent{}, fmt.Errorf("paymentstest: decode webhook payload: %w", err)
	}
	return payments.WebhookEvent{Event: decoded.Event, Payload: decoded.Payload}, nil
}

// SignWebhook signs rawBody the way ParseWebhook expects, so a test can
// build a webhook request without any HTTP.
func (f *FakeGateway) SignWebhook(rawBody []byte) string {
	return fakeSignature(string(rawBody))
}

func fakeSignature(message string) string {
	mac := hmac.New(sha256.New, []byte(fakeGatewaySecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
