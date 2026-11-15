package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/memory-wars/backend/internal/config"
	"github.com/memory-wars/backend/internal/database"
	"github.com/memory-wars/backend/internal/handlers"
	"github.com/memory-wars/backend/internal/middleware"
	"github.com/memory-wars/backend/internal/queue"
	"github.com/memory-wars/backend/internal/services"
	"github.com/memory-wars/backend/internal/worker"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	database.StartKeepalive(ctx, db, 1*time.Minute)

	var sqsClient *queue.SQSClient
	if cfg.SQSQueueURL != "" {
		sqsClient, err = queue.NewSQSClient(cfg.AWSRegion, cfg.SQSQueueURL)
		if err != nil {
			log.Printf("Warning: failed to init SQS client: %v", err)
		} else {
			emailWorker := worker.NewEmailWorker(sqsClient, cfg)
			go emailWorker.Start(ctx)
		}
	} else {
		log.Println("SQS_QUEUE_URL not set, running without email queue (codes printed to console)")
	}

	authService := services.NewAuthService(db, cfg, sqsClient)
	authHandler := handlers.NewAuthHandler(authService)
	creatorService := services.NewCreatorService(db)
	creatorHandler := handlers.NewCreatorHandler(creatorService)

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
				protected.POST("/verify-code", authHandler.VerifyCode)
				protected.GET("/verify-email/status", authHandler.CheckEmailVerification)
				protected.POST("/verify-email/resend", authHandler.ResendVerificationEmail)
				protected.POST("/logout", authHandler.Logout)
				protected.POST("/set-username", authHandler.SetUsername)
				protected.GET("/check-username", authHandler.CheckUsername)
				protected.POST("/set-avatar", authHandler.SetAvatar)
				protected.GET("/profile", authHandler.GetProfile)
				protected.POST("/creator/apply", creatorHandler.Apply)
				protected.GET("/creator/status", creatorHandler.GetApplicationStatus)
			}
		}
	}

	log.Printf("Starting Memory Wars backend on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
