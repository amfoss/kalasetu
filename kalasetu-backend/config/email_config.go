package config

import "os"

type EmailConfig struct {
	GmailClientID     string
	GmailClientSecret string
	GmailRefreshToken string
	FromEmail         string
	FromName          string
}

func LoadEmailConfig() *EmailConfig {
	fromEmail := os.Getenv("GMAIL_USER")

	return &EmailConfig{
		GmailClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		GmailClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		GmailRefreshToken: os.Getenv("GMAIL_REFRESH_TOKEN"),
		FromEmail:         fromEmail,
		FromName:          "KalaSetu",
	}
}

func (c *EmailConfig) IsOAuthConfigured() bool {
	return c != nil &&
		c.GmailClientID != "" &&
		c.GmailClientSecret != "" &&
		c.GmailRefreshToken != ""
}
