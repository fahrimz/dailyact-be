package main

import (
	"dailyact/config"
	"dailyact/handlers"
	"dailyact/middleware"
	"dailyact/seeds"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	mobileAuthHandler := handlers.NewMobileAuthHandler(db)
	authMiddleware := middleware.NewAuthMiddleware(db)

	// Initialize router
	r := gin.Default()

	// Add CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.GET("/me", authMiddleware.RequireAuth(), authHandler.GetMe)
		auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)

		// Mobile auth routes
		auth.POST("/google/verify", mobileAuthHandler.VerifyGoogleToken)
	}

	// User management routes (admin only)
	users := r.Group("/users", authMiddleware.RequireAuth(), authMiddleware.RequireSuperAdmin())
	{
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUserByID)
	}

	// Category routes
	categories := r.Group("/categories")
	{
		categories.GET("", handler.GetCategories)
		categories.POST("", authMiddleware.RequireAuth(), authMiddleware.RequireAdmin(), handler.CreateCategory)
		categories.PUT("/:id", authMiddleware.RequireAuth(), authMiddleware.RequireAdmin(), handler.UpdateCategory)
		categories.DELETE("/:id", authMiddleware.RequireAuth(), authMiddleware.RequireAdmin(), handler.DeleteCategory)
	}

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
