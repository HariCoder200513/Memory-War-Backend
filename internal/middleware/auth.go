package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/utils"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid authorization format"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		if claims.Role != "" {
			c.Set("user_role", claims.Role)
		}
		c.Next()
	}
}

func CreatorMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if role == "" {
			userID := c.GetString("user_id")
			_ = db.QueryRow("SELECT COALESCE(role, 'player') FROM users WHERE id = $1", userID).Scan(&role)
		}
		if role != "creator" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Creator access required"})
			c.Abort()
			return
		}
		c.Set("user_role", role)
		c.Next()
	}
}

func ReviewerMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if role == "" {
			userID := c.GetString("user_id")
			_ = db.QueryRow("SELECT COALESCE(role, 'player') FROM users WHERE id = $1", userID).Scan(&role)
		}
		if role != "reviewer" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Reviewer access required"})
			c.Abort()
			return
		}
		c.Set("user_role", role)
		c.Next()
	}
}
