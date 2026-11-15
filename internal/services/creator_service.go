package services

import (
	"database/sql"
	"fmt"

	"github.com/memory-wars/backend/internal/models"
)

type CreatorService struct {
	db *sql.DB
}

func NewCreatorService(db *sql.DB) *CreatorService {
	return &CreatorService{db: db}
}

func (s *CreatorService) Apply(userID, message string) error {
	_, err := s.db.Exec(
		`INSERT INTO creator_applications (user_id, message)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET message = $2, status = 'pending', updated_at = NOW()`,
		userID, message,
	)
	if err != nil {
		return fmt.Errorf("failed to submit application: %w", err)
	}
	return nil
}

func (s *CreatorService) GetApplication(userID string) (*models.CreatorApplication, error) {
	var app models.CreatorApplication
	err := s.db.QueryRow(
		`SELECT id, user_id, message, status, reviewed_at, created_at, updated_at
		 FROM creator_applications WHERE user_id = $1`,
		userID,
	).Scan(&app.ID, &app.UserID, &app.Message, &app.Status, &app.ReviewedAt, &app.CreatedAt, &app.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	return &app, nil
}

func (s *CreatorService) ReviewApplication(applicationUserID, status string) error {
	if status != "approved" && status != "rejected" {
		return fmt.Errorf("invalid status: must be approved or rejected")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE creator_applications SET status = $1, reviewed_at = NOW(), updated_at = NOW()
		 WHERE user_id = $2`,
		status, applicationUserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	if status == "approved" {
		_, err = tx.Exec(
			"UPDATE users SET role = 'creator', updated_at = NOW() WHERE id = $1",
			applicationUserID,
		)
		if err != nil {
			return fmt.Errorf("failed to update user role: %w", err)
		}
	}

	return tx.Commit()
}

func (s *CreatorService) ListPendingApplications() ([]models.CreatorApplication, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, message, status, reviewed_at, created_at, updated_at
		 FROM creator_applications WHERE status = 'pending' ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	defer rows.Close()

	var apps []models.CreatorApplication
	for rows.Next() {
		var app models.CreatorApplication
		if err := rows.Scan(&app.ID, &app.UserID, &app.Message, &app.Status, &app.ReviewedAt, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}
		apps = append(apps, app)
	}
	return apps, nil
}
