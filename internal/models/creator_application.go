package models

import "time"

type CreatorApplication struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  *string   `json:"username"`
	Avatar    string    `json:"avatar"`
	Role      string    `json:"role"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateApplicationRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

type ApplicationsListResponse struct {
	Applications []CreatorApplication `json:"applications"`
}
