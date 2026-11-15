package models

import "time"

type User struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Username      *string   `json:"username" db:"username"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	GoogleID      *string   `json:"-" db:"google_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GoogleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyCodeRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username"`
	Avatar        string  `json:"avatar"`
	EmailVerified bool    `json:"email_verified"`
}

type SetUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=16"`
}

type SetAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required"`
}

type CheckUsernameResponse struct {
	Available bool `json:"available"`
}

type VerifyEmailResponse struct {
	Verified bool `json:"verified"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
