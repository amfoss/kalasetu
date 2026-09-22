package config

import "os"

// defaultRazorpayBaseURL is Razorpay's production API host.
const defaultRazorpayBaseURL = "https://api.razorpay.com"

// RazorpayConfig holds the settings needed to talk to Razorpay.
type RazorpayConfig struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
	// BaseURL is the Razorpay API host. It defaults to the production host
	// and is overridden in tests to point at a stub server.
	BaseURL string
}

// LoadRazorpayConfig reads Razorpay configuration from environment
// variables. It follows the same pattern as LoadDBConfig.
func LoadRazorpayConfig() *RazorpayConfig {
	baseURL := os.Getenv("RAZORPAY_BASE_URL")
	if baseURL == "" {
		baseURL = defaultRazorpayBaseURL
	}

	return &RazorpayConfig{
		KeyID:         os.Getenv("RAZORPAY_KEY_ID"),
		KeySecret:     os.Getenv("RAZORPAY_KEY_SECRET"),
		WebhookSecret: os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
		BaseURL:       baseURL,
	}
}

// IsConfigured reports whether credentials sufficient to call the Razorpay
// API were provided.
func (c *RazorpayConfig) IsConfigured() bool {
	return c != nil && c.KeyID != "" && c.KeySecret != ""
}
