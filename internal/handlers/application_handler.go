package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/services"
)

type ApplicationHandler struct {
	appService *services.ApplicationService
}

func NewApplicationHandler(appService *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	app, err := h.appService.CreateApplication(userID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "create_failed", Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, app)
}

func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	apps, err := h.appService.ListApplications(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "list_failed", Message: err.Error()})
		return
	}
	if apps == nil {
		apps = []models.CreatorApplication{}
	}
	c.JSON(http.StatusOK, models.ApplicationsListResponse{Applications: apps})
}

func (h *ApplicationHandler) ApproveApplication(c *gin.Context) {
	appID := c.Param("id")
	app, err := h.appService.ApproveApplication(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "approve_failed", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *ApplicationHandler) RejectApplication(c *gin.Context) {
	appID := c.Param("id")
	app, err := h.appService.RejectApplication(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "reject_failed", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, app)
}
