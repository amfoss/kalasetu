package routes

import (
	"kalasetu/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterWebhookRoutes wires gateway webhook endpoints. They carry no auth
// middleware of their own - each authenticates by its payload's signature
// instead, verified inside the handler.
func RegisterWebhookRoutes(router *gin.RouterGroup, razorpayWebhookHandler *handlers.RazorpayWebhookHandler) {
	webhookGroup := router.Group("/webhooks")
	{
		webhookGroup.POST("/razorpay", razorpayWebhookHandler.Handle)
	}
}
