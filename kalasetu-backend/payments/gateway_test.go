package payments

import (
	"context"
	"errors"
	"testing"
)

func TestUnconfiguredGatewayFailsEveryCall(t *testing.T) {
	gw := NewUnconfiguredGateway()
	ctx := context.Background()

	if _, err := gw.CreatePaymentOrder(ctx, CreateOrderRequest{}); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("CreatePaymentOrder: got err %v, want ErrNotConfigured", err)
	}
	if err := gw.VerifyConfirmation(ctx, ConfirmationRequest{}); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("VerifyConfirmation: got err %v, want ErrNotConfigured", err)
	}
	if _, err := gw.Refund(ctx, RefundOrderRequest{}); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("Refund: got err %v, want ErrNotConfigured", err)
	}
	if _, err := gw.ParseWebhook(ctx, []byte("{}"), "sig"); !errors.Is(err, ErrNotConfigured) {
		t.Errorf("ParseWebhook: got err %v, want ErrNotConfigured", err)
	}
}
