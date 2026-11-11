package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "registration_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "auth_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	// TODO: implement Google ID token verification and user creation/login
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: "not_implemented", Message: "Google auth not yet implemented"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	if err := h.authService.ForgotPassword(req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "If an account exists with that email, a reset link has been sent."})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	if err := h.authService.ResetPassword(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "reset_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Password reset successfully"})
}

func (h *AuthHandler) CheckEmailVerification(c *gin.Context) {
	userID := c.GetString("user_id")
	verified, err := h.authService.CheckEmailVerification(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.VerifyEmailResponse{Verified: verified})
}

func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.authService.ResendVerificationEmail(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Verification email sent"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "refresh_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Stateless JWT — client simply discards tokens
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Logged out successfully"})
}
