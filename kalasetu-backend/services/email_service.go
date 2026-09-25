package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"kalasetu/config"
)

type EmailService interface {
	SendOTPEmail(ctx context.Context, toEmail, otp, purpose string) error
}

type emailService struct {
	cfg *config.EmailConfig
}

func NewEmailService(cfg *config.EmailConfig) EmailService {
	return &emailService{cfg: cfg}
}

func (s *emailService) SendOTPEmail(ctx context.Context, toEmail, otp, purpose string) error {
	subject := "Your Verification Code"
	body := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n\r\n"+
		"Your KalaSetu OTP for %s is: %s\n\n"+
		"This code is valid for 10 minutes. Do not share it with anyone.",
		s.cfg.FromName, s.cfg.FromEmail, toEmail, subject, purpose, otp)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	err := smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{strings.TrimSpace(toEmail)}, []byte(body))
	if err != nil {
		log.Printf("Failed to send email via SMTP: %v. Fallback OTP: %s", err, otp)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
