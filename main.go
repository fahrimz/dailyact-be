package main

import (
	"dailyact/config"
	"dailyact/handlers"
	"dailyact/seeds"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	// Initialize database
	db := config.InitDB()

	// Seed categories
	if err := seeds.SeedCategories(db); err != nil {
		log.Fatal("Failed to seed categories:", err)
	}

	// Initialize router
	r := gin.Default()

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Category routes
	r.POST("/categories", h.CreateCategory)
	r.GET("/categories", h.GetCategories)

	// Activity routes
	r.POST("/activities", h.CreateActivity)
	r.GET("/activities", h.GetActivities)
	r.GET("/activities/:id", h.GetActivityByID)
	r.PUT("/activities/:id", h.UpdateActivity)
	r.DELETE("/activities/:id", h.DeleteActivity)

	// Start server
	r.Run(":8080")
}
