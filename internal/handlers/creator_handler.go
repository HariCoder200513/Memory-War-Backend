package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/services"
)

type CreatorHandler struct {
	creatorService *services.CreatorService
}

func NewCreatorHandler(creatorService *services.CreatorService) *CreatorHandler {
	return &CreatorHandler{creatorService: creatorService}
}

func (h *CreatorHandler) Apply(c *gin.Context) {
	var req models.ApplyCreatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Message must be at least 10 characters"})
		return
	}

	userID := c.GetString("user_id")
	if err := h.creatorService.Apply(userID, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Application submitted successfully"})
}

func (h *CreatorHandler) GetApplicationStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	app, err := h.creatorService.GetApplication(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	if app == nil {
		c.JSON(http.StatusOK, models.CreatorApplicationResponse{Status: "none"})
		return
	}

	c.JSON(http.StatusOK, models.CreatorApplicationResponse{Status: app.Status, Message: app.Message})
}

func (h *CreatorHandler) ListPending(c *gin.Context) {
	apps, err := h.creatorService.ListPendingApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "server_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps})
}

func (h *CreatorHandler) Review(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "user_id and status are required"})
		return
	}

	if err := h.creatorService.ReviewApplication(req.UserID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "review_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Application " + req.Status})
}
