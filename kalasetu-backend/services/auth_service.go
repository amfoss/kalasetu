package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"kalasetu/models"
	"kalasetu/repos"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")

	ErrInvalidPassword    = errors.New("current password is incorrect")

	ErrSamePassword       = errors.New("new password cannot be same as the current one")
	ErrInvalidOTP         = errors.New("invalid or expired verification code")
	ErrEmailNotVerified   = errors.New("email is not verified, please verify OTP first")
)

const otpValidityDuration = 10 * time.Minute

type AuthService interface {
	Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error)
	Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error)
	RefreshToken(ctx context.Context, token string) (*models.TokenResponse, error)
	ChangePassword(ctx context.Context, userID int, input models.ChangePasswordInput) error
	SendOTP(ctx context.Context, input models.SendOTPInput) error
	VerifyOTP(ctx context.Context, input models.VerifyOTPInput) (bool, error)
}

type authService struct {
	userRepo         repos.UserRepository
	refreshTokenRepo repos.RefreshTokenRepository
	otpRepo          repos.OTPRepository
	emailService     EmailService
}

func NewAuthService(
	userRepo repos.UserRepository,
	refreshTokenRepo repos.RefreshTokenRepository,
	otpRepo repos.OTPRepository,
	emailService EmailService,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		otpRepo:          otpRepo,
		emailService:     emailService,
	}
}

func (s *authService) SendOTP(ctx context.Context, input models.SendOTPInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	purpose := strings.TrimSpace(input.Purpose)
	if purpose == "" {
		purpose = "signup"
	}

	if purpose == "signup" {
		existing, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrUserAlreadyExists
		}
	}

	otp, err := generateNumericOTP()
	if err != nil {
		return fmt.Errorf("failed to generate verification code: %w", err)
	}

	expiresAt := time.Now().Add(otpValidityDuration)
	if s.otpRepo != nil {
		if err := s.otpRepo.CreateOTP(ctx, email, otp, purpose, expiresAt); err != nil {
			return fmt.Errorf("failed to store verification code: %w", err)
		}
	}

	if s.emailService != nil {
		if err := s.emailService.SendOTPEmail(ctx, email, otp, purpose); err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, input models.VerifyOTPInput) (bool, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	otp := strings.TrimSpace(input.OTP)
	purpose := strings.TrimSpace(input.Purpose)
	if purpose == "" {
		purpose = "signup"
	}

	if s.otpRepo == nil {
		return true, nil
	}

	ok, err := s.otpRepo.VerifyAndConsumeOTP(ctx, email, otp, purpose)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, ErrInvalidOTP
	}

	return true, nil
}

func (s *authService) Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Check if user already exists
	existing, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Verify that email OTP was verified recently (within the last 10 minutes)
	if s.otpRepo != nil {
		verified, err := s.otpRepo.IsEmailVerified(ctx, input.Email, "signup", otpValidityDuration)
		if err != nil {
			return nil, err
		}
		if !verified {
			return nil, ErrEmailNotVerified
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
		Name:     input.Name,
	}

	savedUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(ctx, savedUser)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *savedUser,
	}, nil
}

func (s *authService) Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, rawToken string) (*models.TokenResponse, error) {
	// Hash the supplied token to compare with DB
	hashed := hashToken(rawToken)

	// Look up the token in database
	tokenRecord, err := s.refreshTokenRepo.FindByToken(ctx, hashed)
	if err != nil {
		return nil, err
	}
	if tokenRecord == nil || tokenRecord.Revoked || time.Now().After(tokenRecord.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	// Verify user still exists
	user, err := s.userRepo.FindByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	// Revoke the old refresh token (refresh token rotation)
	if err := s.refreshTokenRepo.Revoke(ctx, hashed); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Generate new access and refresh tokens
	accessToken, newRawRefreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRawRefreshToken,
	}, nil
}



func (s *authService) generateTokens(ctx context.Context, user *models.User) (string, string, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Generate a secure random string for the refresh token
	rawRefreshToken, err := generateRandomToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	hashed := hashToken(rawRefreshToken)

	// Save the hashed refresh token to the database
	expiresAt := time.Now().Add(time.Hour * 24 * 7) // Refresh token expires in 7 days
	rfModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     hashed,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.refreshTokenRepo.Create(ctx, rfModel); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, rawRefreshToken, nil
}

func (s *authService) generateAccessToken(user *models.User) (string, error) {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		return "", errors.New("JWT_ACCESS_SECRET env variable not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), // Access token expires in 15 minutes
	})

	return token.SignedString([]byte(accessSecret))
}

func (s *authService) ChangePassword(ctx context.Context, userID int, input models.ChangePasswordInput) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.NewPassword)) == nil {
		return ErrSamePassword
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.userRepo.UpdatePassword(ctx, userID, string(newHashedPassword))
	if err != nil {
		return err
	}

	if err := s.refreshTokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return err
	}

	return nil
}

func generateNumericOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// Helper: Generate secure random token
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Helper: Hash token using SHA-256
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
