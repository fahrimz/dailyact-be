package main

import (
	"dailyact/config"
	"dailyact/utils"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

// OldActivity represents the activity model before encryption
type OldActivity struct {
	ID          uint   `gorm:"primaryKey"`
	Description string `gorm:"not null"`
	Notes       string
	// Other fields remain the same but are not needed for migration
}

// NewActivity represents the activity model with encrypted fields
type NewActivity struct {
	ID          uint   `gorm:"primaryKey"`
	Description []byte `gorm:"type:text;not null"`
	Notes       []byte `gorm:"type:text"`
	// Other fields remain the same but are not needed for migration
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create encryption service
	encryptionService, err := utils.NewEncryptionService()
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	// Connect to database
	db := config.InitDB()

	// Create backup table
	backupTableName := fmt.Sprintf("activities_backup_%s", time.Now().Format("20060102_150405"))
	log.Printf("Creating backup table: %s", backupTableName)
	if err := db.Exec(fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM activities", backupTableName)).Error; err != nil {
		log.Fatalf("Failed to create backup: %v", err)
	}
	log.Printf("Backup created successfully")

	// Fetch all activities
	var activities []OldActivity
	if err := db.Table("activities").Select("id", "description", "notes").Find(&activities).Error; err != nil {
		log.Fatalf("Failed to fetch activities: %v", err)
	}

	log.Printf("Found %d activities to migrate", len(activities))

	// Process each activity
	successCount := 0
	for _, activity := range activities {
		// Encrypt fields
		descriptionEncrypted, err := encryptionService.Encrypt(activity.Description)
		if err != nil {
			log.Printf("Error encrypting description for activity %d: %v", activity.ID, err)
			continue
		}

		var notesEncrypted string
		if activity.Notes != "" {
			notesEncrypted, err = encryptionService.Encrypt(activity.Notes)
			if err != nil {
				log.Printf("Error encrypting notes for activity %d: %v", activity.ID, err)
				continue
			}
		}

		// Update record with encrypted data
		if err := db.Exec("UPDATE activities SET description = ?, notes = ? WHERE id = ?",
			descriptionEncrypted, notesEncrypted, activity.ID).Error; err != nil {
			log.Printf("Failed to update activity %d: %v", activity.ID, err)
			continue
		}

		successCount++
	}

	log.Printf("Migration completed. %d/%d activities encrypted successfully", successCount, len(activities))
	log.Printf("Backup table name: %s", backupTableName)
	log.Printf("If everything is working correctly, you may delete the backup table after thorough verification")
}
