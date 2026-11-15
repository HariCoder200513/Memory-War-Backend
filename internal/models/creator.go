package models

import "time"

type CreatorApplication struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ApplyCreatorRequest struct {
	Message string `json:"message" binding:"required,min=10"`
}

type CreatorApplicationResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
