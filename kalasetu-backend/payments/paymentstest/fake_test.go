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

func TestFakeRefundsOnlyOncePerReference(t *testing.T) {
	f := paymentstest.NewFake()
	ctx := context.Background()

	res, err := f.Charge(ctx, payments.ChargeRequest{Amount: 500, Reference: "order_1"})
	if err != nil {
		t.Fatalf("charge: %v", err)
	}
	req := payments.RefundRequest{ChargeID: res.ChargeID, Amount: 500, Reference: "item_1"}

	// A repeat of a refund already accepted succeeds without refunding again,
	// so a cancellation retried after a failed commit cannot pay out twice.
	for i := range 2 {
		if err := f.Refund(ctx, req); err != nil {
			t.Fatalf("refund %d: %v", i+1, err)
		}
	}
	if got := len(f.Refunds()); got != 1 {
		t.Errorf("recorded refunds = %d, want 1", got)
	}

	// A different reference is a different refund.
	req.Reference = "item_2"
	if err := f.Refund(ctx, req); err != nil {
		t.Fatalf("second item: %v", err)
	}
	if got := len(f.Refunds()); got != 2 {
		t.Errorf("recorded refunds = %d, want 2", got)
	}
}
