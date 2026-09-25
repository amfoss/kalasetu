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
	"os"
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

)

type AuthService interface {
	Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error)
	Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error)
	RefreshToken(ctx context.Context, token string) (*models.TokenResponse, error)
	ChangePassword(ctx context.Context, userID int, input models.ChangePasswordInput) error
}

type authService struct {
	userRepo         repos.UserRepository
	refreshTokenRepo repos.RefreshTokenRepository
}

func NewAuthService(userRepo repos.UserRepository, refreshTokenRepo repos.RefreshTokenRepository) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *authService) Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error) {
	// Check if user already exists
	existing, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
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
