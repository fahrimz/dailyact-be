package main

import (
	"dailyact/config"
	"dailyact/utils"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Check for encryption key
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Fatal("Error: ENCRYPTION_KEY not found in environment variables. Please generate a key and add it to your .env file.")
	}

	// Initialize encryption service
	encryptionService, err := utils.NewEncryptionService()
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	// Connect to database
	db := config.InitDB()

	// Create backup table
	timestamp := time.Now().Format("20060102_150405")
	backupTableName := fmt.Sprintf("activities_backup_%s", timestamp)
	log.Printf("Creating backup table: %s", backupTableName)

	if err := db.Exec(fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM activities", backupTableName)).Error; err != nil {
		log.Fatalf("Failed to create backup: %v", err)
	}
	log.Printf("✓ Backup created successfully")

	// Fetch all activities
	type ActivityData struct {
		ID          uint
		Description string
		Notes       string
	}

	var activities []ActivityData
	if err := db.Table("activities").Select("id, description, notes").Find(&activities).Error; err != nil {
		log.Fatalf("Failed to fetch activities: %v", err)
	}

	log.Printf("Found %d activities to encrypt", len(activities))

	// Process each activity
	successCount := 0
	errorCount := 0

	for _, activity := range activities {
		// Encrypt fields
		descriptionEncrypted, err := encryptionService.Encrypt(activity.Description)
		if err != nil {
			log.Printf("Error encrypting description for activity %d: %v", activity.ID, err)
			errorCount++
			continue
		}

		var notesEncrypted string
		if activity.Notes != "" {
			notesEncrypted, err = encryptionService.Encrypt(activity.Notes)
			if err != nil {
				log.Printf("Error encrypting notes for activity %d: %v", activity.ID, err)
				errorCount++
				continue
			}
		}

		// Update record with encrypted data
		if err := db.Exec("UPDATE activities SET description = ?, notes = ? WHERE id = ?",
			descriptionEncrypted, notesEncrypted, activity.ID).Error; err != nil {
			log.Printf("Failed to update activity %d: %v", activity.ID, err)
			errorCount++
			continue
		}

		successCount++

		// Show progress for large datasets
		if successCount%100 == 0 {
			log.Printf("Progress: %d/%d activities encrypted", successCount, len(activities))
		}
	}

	log.Printf("\n=========== ENCRYPTION COMPLETE ===========")
	log.Printf("Total activities: %d", len(activities))
	log.Printf("Successfully encrypted: %d", successCount)
	log.Printf("Failed to encrypt: %d", errorCount)
	log.Printf("Backup table: %s", backupTableName)
	log.Printf("==========================================")

	if errorCount > 0 {
		log.Printf("\nWARNING: Some activities could not be encrypted.")
		log.Printf("Review the logs above and fix any issues before proceeding.")
	} else {
		log.Printf("\nAll activities were successfully encrypted!")
	}

	log.Printf("\nIMPORTANT: The backup table %s contains your original unencrypted data.", backupTableName)
	log.Printf("Once you've verified that everything is working correctly, you may want to delete this backup.")
}
