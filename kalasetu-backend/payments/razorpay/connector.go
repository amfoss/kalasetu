// Package razorpay implements payments.Gateway against the Razorpay API
// using the standard library's HTTP client. The Razorpay Go SDK is
// deliberately not taken as a dependency.
package razorpay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"kalasetu/config"
	"kalasetu/payments"
)

const defaultBaseURL = "https://api.razorpay.com"

// ErrInvalidSignature is returned when a confirmation or webhook signature
// does not match the payload it claims to authenticate.
var ErrInvalidSignature = errors.New("razorpay: invalid signature")

// ErrInvalidIdempotencyKey is returned when a refund's idempotency key does
// not meet Razorpay's requirements: at least ten characters drawn from
// alphanumerics, hyphens and underscores.
var ErrInvalidIdempotencyKey = errors.New("razorpay: idempotency key must be at least 10 characters of letters, digits, hyphens or underscores")

var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{10,}$`)

// Connector implements payments.Gateway against the Razorpay API.
type Connector struct {
	httpClient    *http.Client
	baseURL       string
	keyID         string
	keySecret     string
	webhookSecret string
}

// NewConnector builds a Connector from cfg. httpClient may be nil, in which
// case a client with a sane default timeout is used.
func NewConnector(cfg *config.RazorpayConfig, httpClient *http.Client) *Connector {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Connector{
		httpClient:    httpClient,
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		keyID:         cfg.KeyID,
		keySecret:     cfg.KeySecret,
		webhookSecret: cfg.WebhookSecret,
	}
}

var _ payments.Gateway = (*Connector)(nil)

type orderCreateBody struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Receipt  string `json:"receipt,omitempty"`
	// PaymentCapture is sent explicitly (1 for automatic, 0 for manual)
	// because a value supplied through the orders API takes precedence
	// over the dashboard's auto-capture setting.
	PaymentCapture int `json:"payment_capture"`
}

type orderCreateResponse struct {
	ID string `json:"id"`
}

func (c *Connector) CreatePaymentOrder(ctx context.Context, req payments.CreateOrderRequest) (payments.CreateOrderResult, error) {
	capture := 1
	if req.Capture == payments.CaptureManual {
		capture = 0
	}
	body := orderCreateBody{
		Amount:         req.Amount,
		Currency:       req.Currency,
		Receipt:        req.Receipt,
		PaymentCapture: capture,
	}

	status, respBody, err := c.do(ctx, http.MethodPost, "/v1/orders", nil, body)
	if err != nil {
		return payments.CreateOrderResult{}, err
	}
	if status < 200 || status >= 300 {
		return payments.CreateOrderResult{}, fmt.Errorf("razorpay: create order: unexpected status %d: %s", status, respBody)
	}

	var resp orderCreateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return payments.CreateOrderResult{}, fmt.Errorf("razorpay: decode create order response: %w", err)
	}
	return payments.CreateOrderResult{GatewayOrderID: resp.ID}, nil
}

// VerifyConfirmation checks the HMAC-SHA256 signature Razorpay's checkout
// hands back to the Buyer's client, computed over the gateway order id and
// payment id joined by a pipe and keyed with the API secret.
func (c *Connector) VerifyConfirmation(_ context.Context, req payments.ConfirmationRequest) error {
	expected := hmacSHA256(c.keySecret, []byte(req.GatewayOrderID+"|"+req.PaymentID))
	provided, err := hex.DecodeString(req.Signature)
	if err != nil || !hmac.Equal(expected, provided) {
		return ErrInvalidSignature
	}
	return nil
}

type refundCreateBody struct {
	Amount int64 `json:"amount"`
}

type refundResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Refund requests a refund, carrying Razorpay's refund idempotency header.
// A conflict response means the idempotency key was already accepted or is
// in flight, so it is mapped to success rather than an error.
func (c *Connector) Refund(ctx context.Context, req payments.RefundOrderRequest) (payments.RefundOrderResult, error) {
	if !idempotencyKeyPattern.MatchString(req.IdempotencyKey) {
		return payments.RefundOrderResult{}, ErrInvalidIdempotencyKey
	}

	headers := map[string]string{"X-Razorpay-Idempotency": req.IdempotencyKey}
	body := refundCreateBody{Amount: req.Amount}

	status, respBody, err := c.do(ctx, http.MethodPost, "/v1/payments/"+req.PaymentID+"/refund", headers, body)
	if err != nil {
		return payments.RefundOrderResult{}, err
	}

	if status == http.StatusConflict {
		var resp refundResponse
		_ = json.Unmarshal(respBody, &resp) // best-effort: some gateways echo the existing refund on conflict
		return payments.RefundOrderResult{RefundID: resp.ID, Status: resp.Status}, nil
	}
	if status < 200 || status >= 300 {
		return payments.RefundOrderResult{}, fmt.Errorf("razorpay: refund: unexpected status %d: %s", status, respBody)
	}

	var resp refundResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return payments.RefundOrderResult{}, fmt.Errorf("razorpay: decode refund response: %w", err)
	}
	return payments.RefundOrderResult{RefundID: resp.ID, Status: resp.Status}, nil
}

// ParseWebhook verifies the HMAC-SHA256 signature over the raw request body,
// keyed with the webhook secret and compared in constant time. The body is
// never parsed or re-serialised before hashing.
func (c *Connector) ParseWebhook(_ context.Context, rawBody []byte, signature string) (payments.WebhookEvent, error) {
	expected := hmacSHA256(c.webhookSecret, rawBody)
	provided, err := hex.DecodeString(signature)
	if err != nil || !hmac.Equal(expected, provided) {
		return payments.WebhookEvent{}, ErrInvalidSignature
	}

	var decoded struct {
		Event   string          `json:"event"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(rawBody, &decoded); err != nil {
		return payments.WebhookEvent{}, fmt.Errorf("razorpay: decode webhook payload: %w", err)
	}
	return payments.WebhookEvent{Event: decoded.Event, Payload: decoded.Payload}, nil
}

func hmacSHA256(secret string, message []byte) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return mac.Sum(nil)
}

// do sends a JSON request to path and returns the response status and body.
// It does not treat non-2xx statuses as errors - callers interpret those,
// since e.g. a refund conflict is meaningful rather than exceptional.
func (c *Connector) do(ctx context.Context, method, path string, headers map[string]string, body any) (status int, respBody []byte, err error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("razorpay: encode request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, fmt.Errorf("razorpay: build request: %w", err)
	}
	req.SetBasicAuth(c.keyID, c.keySecret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("razorpay: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("razorpay: read response: %w", err)
	}
	return resp.StatusCode, respBody, nil
}
