package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/services"
)

type ReviewerHandler struct {
	gameService *services.GameService
}

func NewReviewerHandler(gameService *services.GameService) *ReviewerHandler {
	return &ReviewerHandler{gameService: gameService}
}

func (h *ReviewerHandler) ListPendingGames(c *gin.Context) {
	status := c.DefaultQuery("status", "pending_review")
	games, err := h.gameService.ListGamesByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "list_failed", Message: err.Error()})
		return
	}
	if games == nil {
		games = []models.Game{}
	}
	c.JSON(http.StatusOK, models.GamesListResponse{Games: games})
}

func (h *ReviewerHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	game, err := h.gameService.GetGameByID(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, game)
}

func (h *ReviewerHandler) ApproveGame(c *gin.Context) {
	gameID := c.Param("id")
	game, err := h.gameService.ReviewGame(gameID, "published")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "approve_failed", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, game)
}

func (h *ReviewerHandler) RejectGame(c *gin.Context) {
	gameID := c.Param("id")
	game, err := h.gameService.ReviewGame(gameID, "rejected")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "reject_failed", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, game)
}
