package paymentstest_test

import (
	"context"
	"errors"
	"testing"

	"kalasetu/payments"
	"kalasetu/payments/paymentstest"
)

func TestFakeGatewayConfirmationPassesVerification(t *testing.T) {
	var gw payments.Gateway = paymentstest.NewFakeGateway()
	fake := paymentstest.NewFakeGateway()
	ctx := context.Background()

	order, err := gw.CreatePaymentOrder(ctx, payments.CreateOrderRequest{Amount: 500, Currency: "INR"})
	if err != nil || order.GatewayOrderID == "" {
		t.Fatalf("CreatePaymentOrder = %+v, %v; want success with an order id", order, err)
	}

	// Confirm is a FakeGateway-specific helper, so exercise it on the
	// concrete type rather than through the payments.Gateway interface.
	order2, err := fake.CreatePaymentOrder(ctx, payments.CreateOrderRequest{Amount: 500, Currency: "INR"})
	if err != nil {
		t.Fatalf("CreatePaymentOrder: %v", err)
	}
	confirmation := fake.Confirm(order2.GatewayOrderID)
	if err := fake.VerifyConfirmation(ctx, confirmation); err != nil {
		t.Fatalf("VerifyConfirmation: %v", err)
	}
}

func TestFakeGatewayVerifyConfirmationRejectsWrongPayment(t *testing.T) {
	fake := paymentstest.NewFakeGateway()
	ctx := context.Background()

	order, err := fake.CreatePaymentOrder(ctx, payments.CreateOrderRequest{Amount: 500})
	if err != nil {
		t.Fatalf("CreatePaymentOrder: %v", err)
	}
	confirmation := fake.Confirm(order.GatewayOrderID)
	confirmation.PaymentID = "someone-elses-payment"

	if err := fake.VerifyConfirmation(ctx, confirmation); err == nil {
		t.Fatal("expected verification to fail for a tampered payment id")
	}
}

func TestFakeGatewayCanBeToldToFail(t *testing.T) {
	fake := paymentstest.NewFakeGateway()
	ctx := context.Background()

	fake.FailCreateOrder(errors.New("gateway down"))
	if _, err := fake.CreatePaymentOrder(ctx, payments.CreateOrderRequest{Amount: 1}); err == nil {
		t.Fatal("expected CreatePaymentOrder to fail")
	}
	fake.FailCreateOrder(nil)

	order, err := fake.CreatePaymentOrder(ctx, payments.CreateOrderRequest{Amount: 1})
	if err != nil {
		t.Fatalf("CreatePaymentOrder should succeed again: %v", err)
	}

	fake.FailRefund(errors.New("refund rejected"))
	if _, err := fake.Refund(ctx, payments.RefundOrderRequest{PaymentID: "pay_1", IdempotencyKey: "refund-key-01"}); err == nil {
		t.Fatal("expected Refund to fail")
	}
	fake.FailRefund(nil)

	if len(fake.Orders()) != 1 {
		t.Fatalf("recorded orders = %d, want 1", len(fake.Orders()))
	}
	_ = order
}

func TestFakeGatewayRefundIsIdempotentPerKey(t *testing.T) {
	fake := paymentstest.NewFakeGateway()
	ctx := context.Background()
	req := payments.RefundOrderRequest{PaymentID: "pay_1", Amount: 500, IdempotencyKey: "refund-key-01"}

	first, err := fake.Refund(ctx, req)
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	second, err := fake.Refund(ctx, req)
	if err != nil {
		t.Fatalf("repeat refund: %v", err)
	}
	if first != second {
		t.Errorf("repeat refund with same key = %+v, want same result %+v", second, first)
	}
	if len(fake.Refunds()) != 1 {
		t.Errorf("recorded refunds = %d, want 1", len(fake.Refunds()))
	}
}

func TestFakeGatewayParseWebhookRoundTrips(t *testing.T) {
	fake := paymentstest.NewFakeGateway()
	ctx := context.Background()
	body := []byte(`{"event":"payment.captured","payload":{"id":"pay_1"}}`)

	sig := fake.SignWebhook(body)
	event, err := fake.ParseWebhook(ctx, body, sig)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	if event.Event != "payment.captured" {
		t.Errorf("Event = %q, want payment.captured", event.Event)
	}

	if _, err := fake.ParseWebhook(ctx, body, "wrong-signature"); err == nil {
		t.Fatal("expected ParseWebhook to reject a bad signature")
	}
}
