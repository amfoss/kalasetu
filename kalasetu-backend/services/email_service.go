package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

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
	if !s.cfg.IsOAuthConfigured() {
		return nil
	}

	return s.sendViaGmailAPI(ctx, toEmail, otp, purpose)
}

func (s *emailService) sendViaGmailAPI(ctx context.Context, toEmail, otp, purpose string) error {
	oauthConfig := &oauth2.Config{
		ClientID:     s.cfg.GmailClientID,
		ClientSecret: s.cfg.GmailClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.send"},
	}

	token := &oauth2.Token{RefreshToken: s.cfg.GmailRefreshToken}
	freshToken, err := oauthConfig.TokenSource(ctx, token).Token()
	if err != nil {
		return fmt.Errorf("gmail oauth2: failed to refresh token: %w", err)
	}

	subject := "Your KalaSetu Verification Code"
	rawMsg := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n"+
			"Your KalaSetu OTP for %s is: %s\r\n\r\nThis code is valid for 10 minutes. Do not share it with anyone.",
		s.cfg.FromName, s.cfg.FromEmail, toEmail, subject, purpose, otp,
	)

	encoded := strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(rawMsg)), "=")

	body, _ := json.Marshal(map[string]string{"raw": encoded})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("gmail api: failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+freshToken.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Gmail API send failed: %v. Fallback OTP: %s", err, otp)
		return fmt.Errorf("failed to send verification email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr map[string]any
		json.NewDecoder(resp.Body).Decode(&apiErr)
		log.Printf("Gmail API error %d: %v. Fallback OTP: %s", resp.StatusCode, apiErr, otp)
		return fmt.Errorf("gmail api returned status %d", resp.StatusCode)
	}

	return nil
}
