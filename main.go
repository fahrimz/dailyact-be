package main

import (
	"dailyact/config"
	"dailyact/handlers"
	"dailyact/middleware"
	"dailyact/seeds"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db := config.InitDB()

	// Seed categories
	if err := seeds.SeedCategories(db); err != nil {
		log.Println(err)
	}

	// Initialize handlers and middleware
	handler := handlers.NewHandler(db)
	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	authMiddleware := middleware.NewAuthMiddleware(db)

	// Initialize router
	r := gin.Default()

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.GET("/me", authMiddleware.RequireAuth(), authHandler.GetMe)
	}

	// User management routes (admin only)
	users := r.Group("/users", authMiddleware.RequireAuth(), authMiddleware.RequireAdmin())
	{
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUserByID)
	}

	// Category routes
	r.POST("/categories", authMiddleware.RequireAuth(), authMiddleware.RequireAdmin(), handler.CreateCategory)
	r.GET("/categories", handler.GetCategories)

	// Activity routes
	// Activities routes (protected)
	activities := r.Group("/activities", authMiddleware.RequireAuth())
	{
		activities.POST("", handler.CreateActivity)
		activities.GET("", handler.GetActivities)
		activities.GET("/:id", authMiddleware.RequireOwnershipOrAdmin(), handler.GetActivityByID)
		activities.PUT("/:id", authMiddleware.RequireOwnershipOrAdmin(), handler.UpdateActivity)
		activities.DELETE("/:id", authMiddleware.RequireOwnershipOrAdmin(), handler.DeleteActivity)
	}

	// Start server
	r.Run(":8080")
}
