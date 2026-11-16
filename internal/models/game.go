package models

import (
	"time"

	"github.com/lib/pq"
)

type Game struct {
	ID               string         `json:"id" db:"id"`
	CreatorID        string         `json:"creator_id" db:"creator_id"`
	Name             string         `json:"name" db:"name"`
	Description      string         `json:"description" db:"description"`
	Category         string         `json:"category" db:"category"`
	GridRows         int            `json:"grid_rows" db:"grid_rows"`
	GridColumns      int            `json:"grid_columns" db:"grid_columns"`
	CardIcons        pq.StringArray `json:"card_icons" db:"card_icons"`
	CardBack         string         `json:"card_back" db:"card_back"`
	TimeLimitSeconds int            `json:"time_limit_seconds" db:"time_limit_seconds"`
	MoveLimit        int            `json:"move_limit" db:"move_limit"`
	MismatchDelayMs  int            `json:"mismatch_delay_ms" db:"mismatch_delay_ms"`
	AllowHints       bool           `json:"allow_hints" db:"allow_hints"`
	Status           string         `json:"status" db:"status"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateGameRequest struct {
	Name             string   `json:"name" binding:"required,max=30"`
	Description      string   `json:"description" binding:"max=120"`
	Category         string   `json:"category" binding:"required"`
	GridRows         int      `json:"grid_rows" binding:"required,min=3,max=10"`
	GridColumns      int      `json:"grid_columns" binding:"required,min=3,max=10"`
	CardIcons        []string `json:"card_icons"`
	CardBack         string   `json:"card_back"`
	TimeLimitSeconds int      `json:"time_limit_seconds"`
	MoveLimit        int      `json:"move_limit"`
	MismatchDelayMs  int      `json:"mismatch_delay_ms"`
	AllowHints       *bool    `json:"allow_hints"`
}

type UpdateGameRequest struct {
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	Category         *string  `json:"category"`
	GridRows         *int     `json:"grid_rows"`
	GridColumns      *int     `json:"grid_columns"`
	CardIcons        []string `json:"card_icons"`
	CardBack         *string  `json:"card_back"`
	TimeLimitSeconds *int     `json:"time_limit_seconds"`
	MoveLimit        *int     `json:"move_limit"`
	MismatchDelayMs  *int     `json:"mismatch_delay_ms"`
	AllowHints       *bool    `json:"allow_hints"`
}

type GamesListResponse struct {
	Games []Game `json:"games"`
}
