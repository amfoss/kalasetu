package razorpay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kalasetu/config"
	"kalasetu/payments"
)

func testConnector(t *testing.T, handler http.HandlerFunc) (*Connector, *httptest.Server) {
	t.Helper()
	if handler == nil {
		handler = func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected HTTP request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	cfg := &config.RazorpayConfig{
		KeyID:         "rzp_test_key",
		KeySecret:     "test_secret",
		WebhookSecret: "webhook_secret",
		BaseURL:       server.URL,
	}
	return NewConnector(cfg, server.Client()), server
}

func sign(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestCreatePaymentOrderSendsAmountAndExplicitCaptureMode(t *testing.T) {
	var gotBody map[string]any
	var gotAuth string

	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/orders" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "order_abc123"})
	})

	result, err := connector.CreatePaymentOrder(context.Background(), payments.CreateOrderRequest{
		Amount:   50000,
		Currency: "INR",
		Receipt:  "order-42",
		Capture:  payments.CaptureManual,
	})
	if err != nil {
		t.Fatalf("CreatePaymentOrder: %v", err)
	}
	if result.GatewayOrderID != "order_abc123" {
		t.Errorf("GatewayOrderID = %q, want order_abc123", result.GatewayOrderID)
	}
	if gotAuth == "" {
		t.Error("expected basic auth header to be set")
	}
	if amount, _ := gotBody["amount"].(float64); int64(amount) != 50000 {
		t.Errorf("amount = %v, want 50000", gotBody["amount"])
	}
	capture, ok := gotBody["payment_capture"]
	if !ok {
		t.Fatal("payment_capture field missing from request body; capture mode must be explicit")
	}
	if capture.(float64) != 0 {
		t.Errorf("payment_capture = %v, want 0 for manual capture", capture)
	}
}

func TestCreatePaymentOrderAutomaticCaptureSendsOne(t *testing.T) {
	var gotBody map[string]any

	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "order_1"})
	})

	_, err := connector.CreatePaymentOrder(context.Background(), payments.CreateOrderRequest{
		Amount:   1000,
		Currency: "INR",
		Capture:  payments.CaptureAutomatic,
	})
	if err != nil {
		t.Fatalf("CreatePaymentOrder: %v", err)
	}
	if capture := gotBody["payment_capture"].(float64); capture != 1 {
		t.Errorf("payment_capture = %v, want 1 for automatic capture", capture)
	}
}

func TestVerifyConfirmationAcceptsValidSignature(t *testing.T) {
	connector, _ := testConnector(t, nil)

	sig := sign("test_secret", "order_abc123|pay_xyz789")
	err := connector.VerifyConfirmation(context.Background(), payments.ConfirmationRequest{
		GatewayOrderID: "order_abc123",
		PaymentID:      "pay_xyz789",
		Signature:      sig,
	})
	if err != nil {
		t.Fatalf("VerifyConfirmation: %v", err)
	}
}

func TestVerifyConfirmationRejectsTamperedSignature(t *testing.T) {
	connector, _ := testConnector(t, nil)

	sig := sign("test_secret", "order_abc123|pay_xyz789")
	err := connector.VerifyConfirmation(context.Background(), payments.ConfirmationRequest{
		GatewayOrderID: "order_other",
		PaymentID:      "pay_xyz789",
		Signature:      sig,
	})
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("got err %v, want ErrInvalidSignature", err)
	}
}

func TestVerifyConfirmationRejectsMalformedSignature(t *testing.T) {
	connector, _ := testConnector(t, nil)

	err := connector.VerifyConfirmation(context.Background(), payments.ConfirmationRequest{
		GatewayOrderID: "order_abc123",
		PaymentID:      "pay_xyz789",
		Signature:      "not-hex!!",
	})
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("got err %v, want ErrInvalidSignature", err)
	}
}

func TestRefundSendsIdempotencyHeader(t *testing.T) {
	var gotHeader string

	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/payments/pay_1/refund" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		gotHeader = r.Header.Get("X-Razorpay-Idempotency")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "rfnd_1", "status": "processed"})
	})

	_, err := connector.Refund(context.Background(), payments.RefundOrderRequest{
		PaymentID:      "pay_1",
		Amount:         500,
		IdempotencyKey: "refund-key-001",
	})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if gotHeader != "refund-key-001" {
		t.Errorf("X-Razorpay-Idempotency = %q, want refund-key-001", gotHeader)
	}
}

func TestRefundRejectsShortIdempotencyKeyWithoutCallingGateway(t *testing.T) {
	called := false
	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	_, err := connector.Refund(context.Background(), payments.RefundOrderRequest{
		PaymentID:      "pay_1",
		Amount:         500,
		IdempotencyKey: "short",
	})
	if !errors.Is(err, ErrInvalidIdempotencyKey) {
		t.Fatalf("got err %v, want ErrInvalidIdempotencyKey", err)
	}
	if called {
		t.Error("gateway should not be called with an invalid idempotency key")
	}
}

func TestRefundRejectsIdempotencyKeyWithInvalidCharacters(t *testing.T) {
	connector, _ := testConnector(t, nil)

	_, err := connector.Refund(context.Background(), payments.RefundOrderRequest{
		PaymentID:      "pay_1",
		Amount:         500,
		IdempotencyKey: "has a space!",
	})
	if !errors.Is(err, ErrInvalidIdempotencyKey) {
		t.Fatalf("got err %v, want ErrInvalidIdempotencyKey", err)
	}
}

func TestRefundConflictIsMappedToSuccess(t *testing.T) {
	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "rfnd_existing", "status": "pending"})
	})

	result, err := connector.Refund(context.Background(), payments.RefundOrderRequest{
		PaymentID:      "pay_1",
		Amount:         500,
		IdempotencyKey: "refund-key-002",
	})
	if err != nil {
		t.Fatalf("Refund: got error %v, want nil (conflict maps to success)", err)
	}
	if result.RefundID != "rfnd_existing" || result.Status != "pending" {
		t.Errorf("result = %+v, want the existing refund's id/status", result)
	}
}

func TestRefundStatusMapping(t *testing.T) {
	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "rfnd_9", "status": "processed"})
	})

	result, err := connector.Refund(context.Background(), payments.RefundOrderRequest{
		PaymentID:      "pay_1",
		Amount:         500,
		IdempotencyKey: "refund-key-003",
	})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if result.RefundID != "rfnd_9" || result.Status != "processed" {
		t.Errorf("result = %+v, want {rfnd_9 processed}", result)
	}
}

func TestRefundErrorStatusReturnsError(t *testing.T) {
	connector, _ := testConnector(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"description":"bad request"}}`))
	})

	_, err := connector.Refund(context.Background(), payments.RefundOrderRequest{
		PaymentID:      "pay_1",
		Amount:         500,
		IdempotencyKey: "refund-key-004",
	})
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
}

func TestParseWebhookVerifiesRawBodySignature(t *testing.T) {
	connector, _ := testConnector(t, nil)

	body := []byte(`{"event":"payment.captured","payload":{"payment":{"entity":{"id":"pay_1"}}}}`)
	sig := sign("webhook_secret", string(body))

	event, err := connector.ParseWebhook(context.Background(), body, sig)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	if event.Event != "payment.captured" {
		t.Errorf("Event = %q, want payment.captured", event.Event)
	}
}

func TestParseWebhookRejectsBadSignature(t *testing.T) {
	connector, _ := testConnector(t, nil)

	body := []byte(`{"event":"payment.captured","payload":{}}`)

	_, err := connector.ParseWebhook(context.Background(), body, sign("wrong_secret", string(body)))
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("got err %v, want ErrInvalidSignature", err)
	}
}

func TestParseWebhookRejectsSignatureOfDifferentBody(t *testing.T) {
	connector, _ := testConnector(t, nil)

	signedBody := []byte(`{"event":"payment.captured","payload":{}}`)
	sig := sign("webhook_secret", string(signedBody))

	tamperedBody := []byte(`{"event":"payment.captured","payload":{"amount":999999}}`)

	_, err := connector.ParseWebhook(context.Background(), tamperedBody, sig)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("got err %v, want ErrInvalidSignature", err)
	}
}
