package routes

import (
	"kalasetu/handlers"
	"kalasetu/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/change-password", middlewares.JWTAuthMiddleware(), authHandler.ChangePassword)
	}
}
