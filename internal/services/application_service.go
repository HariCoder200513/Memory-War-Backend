package services

import (
	"database/sql"
	"fmt"

	"github.com/memory-wars/backend/internal/models"
)

type ApplicationService struct {
	db *sql.DB
}

func NewApplicationService(db *sql.DB) *ApplicationService {
	return &ApplicationService{db: db}
}

func (s *ApplicationService) CreateApplication(userID, reason string) (*models.CreatorApplication, error) {
	var app models.CreatorApplication
	err := s.db.QueryRow(
		`INSERT INTO creator_applications (user_id, reason)
		 SELECT $1, $2
		 WHERE NOT EXISTS (
		     SELECT 1 FROM creator_applications WHERE user_id = $1 AND status = 'pending'
		 )
		 RETURNING id, user_id, reason, status, created_at, updated_at`,
		userID, reason,
	).Scan(&app.ID, &app.UserID, &app.Reason, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("you already have a pending application")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return &app, nil
}

func (s *ApplicationService) ListApplications(status string) ([]models.CreatorApplication, error) {
	rows, err := s.db.Query(
		`SELECT ca.id, ca.user_id, COALESCE(u.username, u.email), COALESCE(u.avatar, 'orbit'),
		        COALESCE(u.role, 'player'), ca.reason, ca.status, ca.created_at, ca.updated_at
		 FROM creator_applications ca
		 JOIN users u ON u.id = ca.user_id
		 WHERE ca.status = $1
		 ORDER BY ca.created_at DESC`, status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	defer rows.Close()

	var apps []models.CreatorApplication
	for rows.Next() {
		var a models.CreatorApplication
		if err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Avatar, &a.Role, &a.Reason, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func (s *ApplicationService) ApproveApplication(appID string) (*models.CreatorApplication, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback()

	var app models.CreatorApplication
	err = tx.QueryRow(
		`UPDATE creator_applications SET status = 'approved', updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'
		 RETURNING id, user_id, reason, status, created_at, updated_at`, appID,
	).Scan(&app.ID, &app.UserID, &app.Reason, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("application not found or already reviewed")
	}

	_, err = tx.Exec(
		"UPDATE users SET role = 'creator', updated_at = NOW() WHERE id = $1", app.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx commit: %w", err)
	}
	return &app, nil
}

func (s *ApplicationService) RejectApplication(appID string) (*models.CreatorApplication, error) {
	var app models.CreatorApplication
	err := s.db.QueryRow(
		`UPDATE creator_applications SET status = 'rejected', updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'
		 RETURNING id, user_id, reason, status, created_at, updated_at`, appID,
	).Scan(&app.ID, &app.UserID, &app.Reason, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("application not found or already reviewed")
	}
	return &app, nil
}
