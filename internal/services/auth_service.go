package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/memory-wars/backend/internal/config"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/utils"
)

type AuthService struct {
	db  *sql.DB
	cfg *config.Config
}

func NewAuthService(db *sql.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = s.db.Exec(
		`INSERT INTO users (id, email, password_hash, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, false, $4, $5)`,
		id, req.Email, hash, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: send verification email

	return s.generateAuthResponse(id, req.Email, false)
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, email_verified FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid email or password")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	return s.generateAuthResponse(user.ID, user.Email, user.EmailVerified)
}

func (s *AuthService) ForgotPassword(email string) error {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	// Always return success to prevent account enumeration
	if !exists {
		return nil
	}

	token, err := generateRandomToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.db.Exec(
		`INSERT INTO password_reset_tokens (token, email, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (email) DO UPDATE SET token = $1, expires_at = $3`,
		token, email, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// TODO: send reset email

	return nil
}

func (s *AuthService) ResetPassword(req *models.ResetPasswordRequest) error {
	var email string
	var expiresAt time.Time
	err := s.db.QueryRow(
		"SELECT email, expires_at FROM password_reset_tokens WHERE token = $1",
		req.Token,
	).Scan(&email, &expiresAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("invalid or expired reset token")
	}
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("reset token has expired")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = s.db.Exec("UPDATE users SET password_hash = $1, updated_at = $2 WHERE email = $3", hash, time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	_, _ = s.db.Exec("DELETE FROM password_reset_tokens WHERE email = $1", email)

	return nil
}

func (s *AuthService) CheckEmailVerification(userID string) (bool, error) {
	var verified bool
	err := s.db.QueryRow("SELECT email_verified FROM users WHERE id = $1", userID).Scan(&verified)
	if err != nil {
		return false, fmt.Errorf("user not found")
	}
	return verified, nil
}

func (s *AuthService) ResendVerificationEmail(userID string) error {
	var email string
	err := s.db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	// TODO: send verification email
	return nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*models.AuthResponse, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	var user models.User
	err = s.db.QueryRow(
		"SELECT id, email, email_verified FROM users WHERE id = $1",
		claims.UserID,
	).Scan(&user.ID, &user.Email, &user.EmailVerified)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.generateAuthResponse(user.ID, user.Email, user.EmailVerified)
}

func (s *AuthService) GetUserByID(userID string) (*models.UserInfo, error) {
	var user models.UserInfo
	err := s.db.QueryRow(
		"SELECT id, email, email_verified FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.EmailVerified)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *AuthService) generateAuthResponse(userID, email string, emailVerified bool) (*models.AuthResponse, error) {
	accessToken, err := utils.GenerateToken(userID, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateToken(userID, s.cfg.JWTSecret, s.cfg.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserInfo{
			ID:            userID,
			Email:         email,
			EmailVerified: emailVerified,
		},
	}, nil
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
