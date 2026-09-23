package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"kalasetu/payments"
	"kalasetu/repos"
	"kalasetu/services"

	"github.com/gin-gonic/gin"
)

// fulfillingEvents are the webhook event types that fulfil a Checkout
// Session. Every other event is acknowledged without any fulfilment
// attempt.
var fulfillingEvents = map[string]bool{
	"payment.captured": true,
	"order.paid":       true,
}

// refundEvents are the webhook event types that advance a cancelled Order
// Item's refund status. Every other event is acknowledged without any
// refund attempt.
var refundEvents = map[string]bool{
	"refund.processed": true,
	"refund.failed":    true,
}

// razorpayWebhookPayload is the shape of the "payment" entity Razorpay
// includes in both a payment.captured and an order.paid webhook's payload.
type razorpayWebhookPayload struct {
	Payment struct {
		Entity struct {
			ID      string `json:"id"`
			OrderID string `json:"order_id"`
		} `json:"entity"`
	} `json:"payment"`
}

// razorpayRefundWebhookPayload is the shape of the "refund" entity Razorpay
// includes in both a refund.processed and a refund.failed webhook's payload.
type razorpayRefundWebhookPayload struct {
	Refund struct {
		Entity struct {
			ID string `json:"id"`
		} `json:"entity"`
	} `json:"refund"`
}

// RazorpayWebhookHandler receives Razorpay's webhook notifications: the
// backstop that fulfils a Checkout Session into an Order even when the
// Buyer's browser never relays a confirmation back. There is no session or
// JWT on this path - the gateway's signature is the only authentication.
type RazorpayWebhookHandler struct {
	gateway payments.Gateway
	orders  services.OrderService
}

func NewRazorpayWebhookHandler(gateway payments.Gateway, orders services.OrderService) *RazorpayWebhookHandler {
	return &RazorpayWebhookHandler{gateway: gateway, orders: orders}
}

// Handle reads the raw request body before anything else touches it, so the
// bytes verified against the signature are exactly the bytes that were
// sent. An invalid or absent signature is rejected before the payload is
// parsed or acted on.
func (h *RazorpayWebhookHandler) Handle(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read request body"})
		return
	}

	event, err := h.gateway.ParseWebhook(c.Request.Context(), rawBody, c.GetHeader("X-Razorpay-Signature"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	if refundEvents[event.Event] {
		h.handleRefundEvent(c, rawBody, event)
		return
	}

	if !fulfillingEvents[event.Event] {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	var payload razorpayWebhookPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed webhook payload"})
		return
	}
	if payload.Payment.Entity.OrderID == "" || payload.Payment.Entity.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook payload is missing payment identifiers"})
		return
	}

	// The event's identifier is derived from the verified raw body rather
	// than trusted from an unauthenticated field, and is stable across
	// Razorpay's retried deliveries of the same event.
	sum := sha256.Sum256(rawBody)
	eventID := hex.EncodeToString(sum[:])

	_, _, err = h.orders.FulfilFromWebhook(c.Request.Context(), eventID, event.Event, payload.Payment.Entity.OrderID, payload.Payment.Entity.ID)
	if err != nil {
		if errors.Is(err, repos.ErrCheckoutSessionNotConfirmable) {
			// The session is cancelled or expired: nothing to fulfil, and
			// retrying will not change that.
			c.JSON(http.StatusOK, gin.H{"status": "ignored"})
			return
		}
		if errors.Is(err, repos.ErrCheckoutSessionNotFound) {
			// No session matches this gateway order id yet. Unlike an
			// expired or cancelled session, this can be transient, so
			// answer with an error rather than "ignored" and let Razorpay
			// retry the delivery.
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no checkout session for this gateway order id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fulfil checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleRefundEvent advances the refund status of the Order Item a
// refund.processed or refund.failed notification refers to. An unknown
// refund identifier is acknowledged rather than treated as an error: it may
// belong to a refund KalaSetu never recorded a refund_id for.
func (h *RazorpayWebhookHandler) handleRefundEvent(c *gin.Context, rawBody []byte, event payments.WebhookEvent) {
	var payload razorpayRefundWebhookPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed webhook payload"})
		return
	}
	if payload.Refund.Entity.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook payload is missing refund identifier"})
		return
	}

	sum := sha256.Sum256(rawBody)
	eventID := hex.EncodeToString(sum[:])

	if _, err := h.orders.AdvanceRefundFromWebhook(c.Request.Context(), eventID, event.Event, payload.Refund.Entity.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not advance refund status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
