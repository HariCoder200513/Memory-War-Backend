package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"github.com/memory-wars/backend/internal/config"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/queue"
	"github.com/memory-wars/backend/internal/utils"
)

type AuthService struct {
	db       *sql.DB
	cfg      *config.Config
	sqsQueue *queue.SQSClient
}

func NewAuthService(db *sql.DB, cfg *config.Config, sqsQueue *queue.SQSClient) *AuthService {
	return &AuthService{db: db, cfg: cfg, sqsQueue: sqsQueue}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New().String()
	now := time.Now()
	code, err := generateSixDigitCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	var insertedID string
	err = s.db.QueryRow(`
		WITH new_user AS (
			INSERT INTO users (id, email, password_hash, email_verified, created_at, updated_at)
			VALUES ($1, $2, $3, false, $4, $5)
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		),
		new_code AS (
			INSERT INTO email_verification_codes (user_id, code, expires_at)
			SELECT id, $6, $7 FROM new_user
			ON CONFLICT (user_id) DO UPDATE SET code = $6, expires_at = $7, created_at = NOW()
		)
		SELECT id FROM new_user`,
		id, req.Email, hash, now, now, code, now.Add(10*time.Minute),
	).Scan(&insertedID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("email already registered")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.sendVerificationAsync(req.Email, code)

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

func (s *AuthService) VerifyCode(userID string, code string) error {
	var updatedID string
	err := s.db.QueryRow(`
		WITH valid AS (
			DELETE FROM email_verification_codes
			WHERE user_id = $1 AND code = $2 AND expires_at > NOW()
			RETURNING user_id
		)
		UPDATE users SET email_verified = true, updated_at = NOW()
		WHERE id = (SELECT user_id FROM valid)
		RETURNING id`,
		userID, code,
	).Scan(&updatedID)

	if err == sql.ErrNoRows {
		var exists bool
		_ = s.db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM email_verification_codes WHERE user_id = $1)",
			userID,
		).Scan(&exists)
		if !exists {
			return fmt.Errorf("no verification code found, please request a new one")
		}
		return fmt.Errorf("invalid or expired verification code")
	}
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (s *AuthService) ForgotPassword(email string) error {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
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

	// TODO: send reset email via SQS

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
	return s.sendVerificationCode(userID, email)
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
		"SELECT id, email, username, avatar, email_verified FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.Avatar, &user.EmailVerified)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *AuthService) SetUsername(userID, username string) error {
	var taken bool
	err := s.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)",
		username, userID,
	).Scan(&taken)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if taken {
		return fmt.Errorf("username already taken")
	}

	_, err = s.db.Exec(
		"UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2",
		username, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}
	return nil
}

func (s *AuthService) SetAvatar(userID, avatar string) error {
	_, err := s.db.Exec(
		"UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2",
		avatar, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to set avatar: %w", err)
	}
	return nil
}

func (s *AuthService) CheckUsername(username string) (bool, error) {
	var taken bool
	err := s.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)",
		username,
	).Scan(&taken)
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return !taken, nil
}

func (s *AuthService) sendVerificationCode(userID, email string) error {
	code, err := generateSixDigitCode()
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = s.db.Exec(
		`INSERT INTO email_verification_codes (user_id, code, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET code = $2, expires_at = $3, created_at = NOW()`,
		userID, code, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	s.sendVerificationAsync(email, code)
	return nil
}

func (s *AuthService) sendVerificationAsync(email, code string) {
	if s.sqsQueue != nil {
		go func() {
			if err := s.sqsQueue.PublishEmailMessage(queue.EmailMessage{
				To:      email,
				Subject: "Memory Wars - Verify your email",
				Code:    code,
				Type:    "verification",
			}); err != nil {
				fmt.Printf("Failed to publish to SQS: %v\n", err)
			}
		}()
		return
	}
	go s.sendEmailDirect(email, "Memory Wars - Verify your email", code)
}

func (s *AuthService) sendEmailDirect(to, subject, code string) {
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
			<h2 style="color: #1a1a2e;">Verify your email</h2>
			<p style="color: #555;">Your Memory Wars verification code is:</p>
			<div style="background: #f4f4f8; border-radius: 12px; padding: 24px; text-align: center; margin: 24px 0;">
				<span style="font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #1a1a2e;">%s</span>
			</div>
			<p style="color: #888; font-size: 13px;">This code expires in 10 minutes. If you didn't create an account, ignore this email.</p>
		</div>
	`, code)

	mime := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.cfg.SMTPFrom, to, subject, body,
	)

	if err := smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, []byte(mime)); err != nil {
		fmt.Printf("Failed to send email to %s: %v\n", to, err)
	} else {
		fmt.Printf("Verification email sent to %s\n", to)
	}
}

func (s *AuthService) generateAuthResponse(userID, email string, emailVerified bool) (*models.AuthResponse, error) {
	var username *string
	var avatar string
	_ = s.db.QueryRow("SELECT username, COALESCE(avatar, 'orbit') FROM users WHERE id = $1", userID).Scan(&username, &avatar)

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
			Username:      username,
			Avatar:        avatar,
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

func generateSixDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
