package handlers

import (
	"errors"
	"kalasetu/middlewares"
	"kalasetu/models"
	"kalasetu/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Helper to set httpOnly cookie for refresh token
func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	c.SetCookie(
		"refresh_token", // name
		token,           // value
		maxAge,          // maxAge in seconds
		"/",             // path
		"",              // domain
		false,           // secure (false for HTTP development on localhost, set true for HTTPS in prod)
		true,            // httpOnly (prevents client-side JS access)
	)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Pass refresh token via httpOnly cookie, and clear it from the JSON response body
	setRefreshTokenCookie(c, res.RefreshToken, 60*60*24*7) // 7 days
	res.RefreshToken = ""

	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Pass refresh token via httpOnly cookie, and clear it from the JSON response body
	setRefreshTokenCookie(c, res.RefreshToken, 60*60*24*7) // 7 days
	res.RefreshToken = ""

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := middlewares.GetUserIDFromContext(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.authService.ChangePassword(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect current password"})
			return
		}
		if errors.Is(err, services.ErrSamePassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token cookie is missing"})
		return
	}

	res, err := h.authService.RefreshToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set the new rotated refresh token in the httpOnly cookie
	setRefreshTokenCookie(c, res.RefreshToken, 60*60*24*7)

	c.JSON(http.StatusOK, res)
}


