package config

import (
	"os"
	"strconv"
	"time"
)

// defaultRazorpayBaseURL is Razorpay's production API host.
const defaultRazorpayBaseURL = "https://api.razorpay.com"

// defaultReservationWindow is how long a Checkout Session holds its Stock
// reservation by default. It must exceed Razorpay's documented 12-minute
// minimum auto-capture window, or a reservation window shorter than it would
// manufacture late payments by construction.
const defaultReservationWindow = 20 * time.Minute

// RazorpayConfig holds the settings needed to talk to Razorpay.
type RazorpayConfig struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
	// BaseURL is the Razorpay API host. It defaults to the production host
	// and is overridden in tests to point at a stub server.
	BaseURL string
	// ReservationWindow is how long Stock stays reserved by a Checkout
	// Session. Configurable via RAZORPAY_RESERVATION_WINDOW_MINUTES.
	ReservationWindow time.Duration
}

// LoadRazorpayConfig reads Razorpay configuration from environment
// variables. It follows the same pattern as LoadDBConfig.
func LoadRazorpayConfig() *RazorpayConfig {
	baseURL := os.Getenv("RAZORPAY_BASE_URL")
	if baseURL == "" {
		baseURL = defaultRazorpayBaseURL
	}

	reservationWindow := defaultReservationWindow
	if v := os.Getenv("RAZORPAY_RESERVATION_WINDOW_MINUTES"); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil && minutes > 0 {
			reservationWindow = time.Duration(minutes) * time.Minute
		}
	}

	return &RazorpayConfig{
		KeyID:             os.Getenv("RAZORPAY_KEY_ID"),
		KeySecret:         os.Getenv("RAZORPAY_KEY_SECRET"),
		WebhookSecret:     os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
		BaseURL:           baseURL,
		ReservationWindow: reservationWindow,
	}
}

// IsConfigured reports whether credentials sufficient to call the Razorpay
// API were provided.
func (c *RazorpayConfig) IsConfigured() bool {
	return c != nil && c.KeyID != "" && c.KeySecret != ""
}
