package models

import "time"

type EmailOTP struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	OTP       string    `json:"otp"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

type SendOTPInput struct {
	Email   string `json:"email" binding:"required,email"`
	Purpose string `json:"purpose"` // e.g. "signup", "reset_password", "login"
}

type VerifyOTPInput struct {
	Email   string `json:"email" binding:"required,email"`
	OTP     string `json:"otp" binding:"required,len=6"`
	Purpose string `json:"purpose"`
}
