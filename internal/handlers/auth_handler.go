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

func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var req models.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	userID := c.GetString("user_id")
	if err := h.authService.VerifyCode(userID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "verification_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Email verified successfully"})
}

func (h *AuthHandler) GoogleAuth(c *gin.Context) {
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
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Logged out successfully"})
}

func (h *AuthHandler) SetUsername(c *gin.Context) {
	var req models.SetUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Username must be 3-16 characters"})
		return
	}

	userID := c.GetString("user_id")
	if err := h.authService.SetUsername(userID, req.Username); err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "username_taken", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Username set successfully"})
}

func (h *AuthHandler) SetAvatar(c *gin.Context) {
	var req models.SetAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Avatar is required"})
		return
	}

	userID := c.GetString("user_id")
	if err := h.authService.SetAvatar(userID, req.Avatar); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Avatar set successfully"})
}

func (h *AuthHandler) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Username is required"})
		return
	}

	available, err := h.authService.CheckUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.CheckUsernameResponse{Available: available})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
