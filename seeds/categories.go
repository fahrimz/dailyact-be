package seeds

import (
	"dailyact/models"
	"gorm.io/gorm"
	"log"
)

func SeedCategories(db *gorm.DB) error {
	// Check if categories already exist
	var count int64
	if err := db.Model(&models.Category{}).Count(&count).Error; err != nil {
		return err
	}

	// Skip if categories already exist
	if count > 0 {
		log.Println("Categories already exist, skipping seed")
		return nil
	}
	categories := []models.Category{
		{
			Name:        "Sleep",
			Description: "Sleep and rest activities",
		},
		{
			Name:        "Fitness",
			Description: "Physical exercise and wellness activities",
		},
		{
			Name:        "Work",
			Description: "Work-related tasks and professional development",
		},
		{
			Name:        "Meal",
			Description: "Food, drinks, and nutrition activities",
		},
		{
			Name:        "Study",
			Description: "Learning and educational activities",
		},
		{
			Name:        "Entertainment",
			Description: "Leisure and recreational activities",
		},
		{
			Name:        "Social",
			Description: "Social interactions and relationships",
		},
	}

	for _, category := range categories {
		if err := db.Create(&category).Error; err != nil {
			log.Printf("Error seeding category %s: %v\n", category.Name, err)
			return err
		}
		log.Printf("Category seeded successfully: %s\n", category.Name)
	}

	return nil
}
