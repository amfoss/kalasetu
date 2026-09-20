package paymentstest_test

import (
	"context"
	"errors"
	"testing"

	"kalasetu/payments"
	"kalasetu/payments/paymentstest"
)

func TestFakeSucceedsByDefault(t *testing.T) {
	var p payments.PaymentProvider = paymentstest.NewFake()
	ctx := context.Background()

	res, err := p.Charge(ctx, payments.ChargeRequest{Amount: 500, Reference: "order-1"})
	if err != nil || res.ChargeID == "" {
		t.Fatalf("charge = %+v, %v; want success with a charge id", res, err)
	}
	if err := p.Refund(ctx, payments.RefundRequest{ChargeID: res.ChargeID, Amount: 500}); err != nil {
		t.Fatalf("refund: %v", err)
	}
}

func TestFakeCanBeToldToFail(t *testing.T) {
	f := paymentstest.NewFake()
	ctx := context.Background()

	f.FailCharges(errors.New("card declined"))
	if _, err := f.Charge(ctx, payments.ChargeRequest{Amount: 1}); err == nil {
		t.Fatal("expected charge to fail")
	}

	f.FailCharges(nil)
	res, err := f.Charge(ctx, payments.ChargeRequest{Amount: 1})
	if err != nil {
		t.Fatalf("charge should succeed again: %v", err)
	}

	f.FailRefunds(errors.New("refund rejected"))
	if err := f.Refund(ctx, payments.RefundRequest{ChargeID: res.ChargeID, Amount: 1}); err == nil {
		t.Fatal("expected refund to fail")
	}

	if got := len(f.Charges()); got != 1 {
		t.Fatalf("recorded successful charges = %d, want 1", got)
	}
}
