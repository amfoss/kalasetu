package paymentstest

import (
	"context"
	"fmt"
	"sync"

	"kalasetu/payments"
)

// Fake is an in-memory PaymentProvider for tests. It succeeds unless told to fail.
type Fake struct {
	mu        sync.Mutex
	chargeErr error
	refundErr error
	nextID    int
	charges   []payments.ChargeRequest
	refunds   []payments.RefundRequest
	refunded  map[string]bool // references already refunded
}

func NewFake() *Fake { return &Fake{refunded: map[string]bool{}} }

// FailCharges makes subsequent charges return err; nil restores success.
func (f *Fake) FailCharges(err error) { f.mu.Lock(); f.chargeErr = err; f.mu.Unlock() }

// FailRefunds makes subsequent refunds return err; nil restores success.
func (f *Fake) FailRefunds(err error) { f.mu.Lock(); f.refundErr = err; f.mu.Unlock() }

// Charges returns the successful charges so far.
func (f *Fake) Charges() []payments.ChargeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]payments.ChargeRequest(nil), f.charges...)
}

// Refunds returns the successful refunds so far.
func (f *Fake) Refunds() []payments.RefundRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]payments.RefundRequest(nil), f.refunds...)
}

func (f *Fake) Charge(_ context.Context, req payments.ChargeRequest) (payments.ChargeResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.chargeErr != nil {
		return payments.ChargeResult{}, f.chargeErr
	}
	f.nextID++
	f.charges = append(f.charges, req)
	return payments.ChargeResult{ChargeID: fmt.Sprintf("fake_charge_%d", f.nextID)}, nil
}

// Refund records the refund, or succeeds without recording a second one when
// req.Reference has already been refunded, the way a real gateway treats a
// repeated idempotency key.
func (f *Fake) Refund(_ context.Context, req payments.RefundRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.refundErr != nil {
		return f.refundErr
	}
	if req.Reference != "" && f.refunded[req.Reference] {
		return nil
	}
	if req.Reference != "" {
		f.refunded[req.Reference] = true
	}
	f.refunds = append(f.refunds, req)
	return nil
}
