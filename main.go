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
	h := handlers.NewHandler(db)
	authHandler := handlers.NewAuthHandler(db)
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

	// Category routes
	r.POST("/categories", authMiddleware.RequireAuth(), authMiddleware.RequireAdmin(), h.CreateCategory)
	r.GET("/categories", h.GetCategories)

	// Activity routes
	// Activities routes (protected)
	activities := r.Group("/activities", authMiddleware.RequireAuth())
	{
		activities.POST("", h.CreateActivity)
		activities.GET("", h.GetActivities)
		activities.GET("/:id", authMiddleware.RequireOwnershipOrAdmin(), h.GetActivityByID)
		activities.PUT("/:id", authMiddleware.RequireOwnershipOrAdmin(), h.UpdateActivity)
		activities.DELETE("/:id", authMiddleware.RequireOwnershipOrAdmin(), h.DeleteActivity)
	}

	// Start server
	r.Run(":8080")
}
