package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/config"
	"github.com/memory-wars/backend/internal/database"
	"github.com/memory-wars/backend/internal/handlers"
	"github.com/memory-wars/backend/internal/middleware"
	"github.com/memory-wars/backend/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	authService := services.NewAuthService(db, cfg)
	authHandler := handlers.NewAuthHandler(authService)

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleAuth)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/refresh", authHandler.RefreshToken)

			protected := auth.Group("")
			protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				protected.GET("/verify-email/status", authHandler.CheckEmailVerification)
				protected.POST("/verify-email/resend", authHandler.ResendVerificationEmail)
				protected.POST("/logout", authHandler.Logout)
			}
		}
	}

	log.Printf("Starting Memory Wars backend on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
