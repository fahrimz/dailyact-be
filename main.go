package main

import (
	"dailyact/config"
	"dailyact/handlers"
	"dailyact/middleware"
	"dailyact/models"
	"dailyact/seeds"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	// Check if in development environment - load .env file only in dev mode
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
	encryptionService, err := models.NewEncryptionService()
	if err != nil {
		log.Println(err)
		return
	}

	handler, _ := handlers.NewHandler(db, encryptionService)
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
		users.PUT("/:id/change_role", userHandler.ChangeRole)
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

	// App feedback routes
	appFeedback := r.Group("/app_feedbacks", authMiddleware.RequireAuth())
	{
		appFeedback.POST("", handler.CreateAppFeedback)
		appFeedback.GET("", authMiddleware.RequireSuperAdmin(), handler.GetAppFeedbacks)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3005"
	}
	r.Run(":" + port) // Default port is 8080
}
