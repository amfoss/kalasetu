package config

import (
	"os"
	"strconv"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func LoadEmailConfig() *EmailConfig {
	port := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			port = parsed
		}
	}

	fromEmail := os.Getenv("SMTP_FROM")
	if fromEmail == "" {
		fromEmail = os.Getenv("SMTP_USER")
	}

	return &EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     port,
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    fromEmail,
		FromName:     "KalaSetu",
	}
}

func (c *EmailConfig) IsConfigured() bool {
	return c != nil && c.SMTPHost != "" && c.SMTPUser != "" && c.SMTPPassword != ""
}
