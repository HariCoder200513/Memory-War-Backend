package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/memory-wars/backend/internal/models"
)

const gameColumns = `id, creator_id, name, description, category, grid_rows, grid_columns,
	card_icons, card_back, time_limit_seconds, move_limit, mismatch_delay_ms, allow_hints,
	status, created_at, updated_at`

type GameService struct {
	db *sql.DB
}

func NewGameService(db *sql.DB) *GameService {
	return &GameService{db: db}
}

func scanGame(row interface{ Scan(dest ...interface{}) error }) (*models.Game, error) {
	var g models.Game
	err := row.Scan(
		&g.ID, &g.CreatorID, &g.Name, &g.Description, &g.Category,
		&g.GridRows, &g.GridColumns, &g.CardIcons, &g.CardBack,
		&g.TimeLimitSeconds, &g.MoveLimit, &g.MismatchDelayMs, &g.AllowHints,
		&g.Status, &g.CreatedAt, &g.UpdatedAt,
	)
	return &g, err
}

func (s *GameService) CreateGame(creatorID string, req *models.CreateGameRequest) (*models.Game, error) {
	id := uuid.New().String()
	now := time.Now()

	cardIcons := pq.StringArray(req.CardIcons)
	if cardIcons == nil {
		cardIcons = pq.StringArray{}
	}
	cardBack := req.CardBack
	if cardBack == "" {
		cardBack = "classic"
	}
	timeLimitSeconds := req.TimeLimitSeconds
	if timeLimitSeconds == 0 {
		timeLimitSeconds = 120
	}
	mismatchDelayMs := req.MismatchDelayMs
	if mismatchDelayMs == 0 {
		mismatchDelayMs = 1000
	}
	allowHints := true
	if req.AllowHints != nil {
		allowHints = *req.AllowHints
	}

	game := &models.Game{
		ID:               id,
		CreatorID:        creatorID,
		Name:             req.Name,
		Description:      req.Description,
		Category:         req.Category,
		GridRows:         req.GridRows,
		GridColumns:      req.GridColumns,
		CardIcons:        cardIcons,
		CardBack:         cardBack,
		TimeLimitSeconds: timeLimitSeconds,
		MoveLimit:        req.MoveLimit,
		MismatchDelayMs:  mismatchDelayMs,
		AllowHints:       allowHints,
		Status:           "draft",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	_, err := s.db.Exec(`
		INSERT INTO games (`+gameColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		game.ID, game.CreatorID, game.Name, game.Description, game.Category,
		game.GridRows, game.GridColumns, game.CardIcons, game.CardBack,
		game.TimeLimitSeconds, game.MoveLimit, game.MismatchDelayMs, game.AllowHints,
		game.Status, game.CreatedAt, game.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	return game, nil
}

func (s *GameService) GetGame(gameID, creatorID string) (*models.Game, error) {
	game, err := scanGame(s.db.QueryRow(
		`SELECT `+gameColumns+` FROM games WHERE id = $1 AND creator_id = $2`,
		gameID, creatorID,
	))
	if err != nil {
		log.Printf("[GetGame] scan failed for gameID=%s creatorID=%s: %v", gameID, creatorID, err)
		return nil, fmt.Errorf("game not found: %w", err)
	}
	return game, nil
}

func (s *GameService) ListGames(creatorID, status string) ([]models.Game, error) {
	query := `SELECT ` + gameColumns + ` FROM games WHERE creator_id = $1`
	args := []interface{}{creatorID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY updated_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, *g)
	}
	return games, nil
}

func (s *GameService) UpdateGame(gameID, creatorID string, req *models.UpdateGameRequest) (*models.Game, error) {
	game, err := s.GetGame(gameID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("game not found")
	}
	if game.Status != "draft" {
		return nil, fmt.Errorf("can only edit draft games")
	}

	if req.Name != nil {
		game.Name = *req.Name
	}
	if req.Description != nil {
		game.Description = *req.Description
	}
	if req.Category != nil {
		game.Category = *req.Category
	}
	if req.GridRows != nil {
		game.GridRows = *req.GridRows
	}
	if req.GridColumns != nil {
		game.GridColumns = *req.GridColumns
	}
	if req.CardIcons != nil {
		game.CardIcons = pq.StringArray(req.CardIcons)
	}
	if req.CardBack != nil {
		game.CardBack = *req.CardBack
	}
	if req.TimeLimitSeconds != nil {
		game.TimeLimitSeconds = *req.TimeLimitSeconds
	}
	if req.MoveLimit != nil {
		game.MoveLimit = *req.MoveLimit
	}
	if req.MismatchDelayMs != nil {
		game.MismatchDelayMs = *req.MismatchDelayMs
	}
	if req.AllowHints != nil {
		game.AllowHints = *req.AllowHints
	}

	updated, err := scanGame(s.db.QueryRow(`
		UPDATE games SET name = $1, description = $2, category = $3, grid_rows = $4, grid_columns = $5,
			card_icons = $6, card_back = $7, time_limit_seconds = $8, move_limit = $9,
			mismatch_delay_ms = $10, allow_hints = $11, updated_at = NOW()
		WHERE id = $12 AND creator_id = $13
		RETURNING `+gameColumns,
		game.Name, game.Description, game.Category, game.GridRows, game.GridColumns,
		game.CardIcons, game.CardBack, game.TimeLimitSeconds, game.MoveLimit,
		game.MismatchDelayMs, game.AllowHints,
		gameID, creatorID,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}
	return updated, nil
}

func (s *GameService) DeleteGame(gameID, creatorID string) error {
	result, err := s.db.Exec(
		"DELETE FROM games WHERE id = $1 AND creator_id = $2 AND status = 'draft'",
		gameID, creatorID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("game not found or not in draft status")
	}
	return nil
}

func (s *GameService) ListGamesByStatus(status string) ([]models.Game, error) {
	rows, err := s.db.Query(
		`SELECT `+gameColumns+` FROM games WHERE status = $1 ORDER BY updated_at DESC`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, *g)
	}
	return games, nil
}

func (s *GameService) ReviewGame(gameID, status string) (*models.Game, error) {
	if status != "published" && status != "rejected" {
		return nil, fmt.Errorf("invalid review status: must be published or rejected")
	}

	game, err := scanGame(s.db.QueryRow(
		`UPDATE games SET status = $1, updated_at = NOW()
		 WHERE id = $2 AND status = 'pending_review'
		 RETURNING `+gameColumns,
		status, gameID,
	))
	if err != nil {
		return nil, fmt.Errorf("game not found or not in pending_review status")
	}
	return game, nil
}

func (s *GameService) GetGameByID(gameID string) (*models.Game, error) {
	game, err := scanGame(s.db.QueryRow(
		`SELECT `+gameColumns+` FROM games WHERE id = $1`, gameID,
	))
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}
	return game, nil
}

func (s *GameService) SubmitGame(gameID, creatorID string) (*models.Game, error) {
	game, err := scanGame(s.db.QueryRow(
		`UPDATE games SET status = 'pending_review', updated_at = NOW()
		 WHERE id = $1 AND creator_id = $2 AND status = 'draft'
		 RETURNING `+gameColumns,
		gameID, creatorID,
	))
	if err != nil {
		return nil, fmt.Errorf("game not found or not in draft status")
	}
	return game, nil
}
