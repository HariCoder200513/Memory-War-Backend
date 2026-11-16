package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/models"
	"github.com/memory-wars/backend/internal/services"
)

type GameHandler struct {
	gameService *services.GameService
}

func NewGameHandler(gameService *services.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

func (h *GameHandler) CreateGame(c *gin.Context) {
	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	creatorID := c.GetString("user_id")
	game, err := h.gameService.CreateGame(creatorID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "create_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, game)
}

func (h *GameHandler) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	creatorID := c.GetString("user_id")

	game, err := h.gameService.GetGame(gameID, creatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}

func (h *GameHandler) ListGames(c *gin.Context) {
	creatorID := c.GetString("user_id")
	status := c.Query("status")

	games, err := h.gameService.ListGames(creatorID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "list_failed", Message: err.Error()})
		return
	}

	if games == nil {
		games = []models.Game{}
	}

	c.JSON(http.StatusOK, models.GamesListResponse{Games: games})
}

func (h *GameHandler) UpdateGame(c *gin.Context) {
	var req models.UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Invalid request body"})
		return
	}

	gameID := c.Param("id")
	creatorID := c.GetString("user_id")

	game, err := h.gameService.UpdateGame(gameID, creatorID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "update_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}

func (h *GameHandler) DeleteGame(c *gin.Context) {
	gameID := c.Param("id")
	creatorID := c.GetString("user_id")

	if err := h.gameService.DeleteGame(gameID, creatorID); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "delete_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Game deleted"})
}

func (h *GameHandler) SubmitGame(c *gin.Context) {
	gameID := c.Param("id")
	creatorID := c.GetString("user_id")

	game, err := h.gameService.SubmitGame(gameID, creatorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "submit_failed", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, game)
}
