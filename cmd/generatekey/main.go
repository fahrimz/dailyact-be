package main

import (
	"dailyact/models"
	"fmt"
)

func main() {
	// Generate a new encryption key
	key, err := models.GenerateKey()
	if err != nil {
		fmt.Printf("Error generating encryption key: %v\n", err)
		return
	}

	// Print the key with instructions
	fmt.Println("\n========== ENCRYPTION KEY ==========")
	fmt.Println(key)
	fmt.Println("===================================")
	fmt.Println("\nAdd this key to your .env file as:")
	fmt.Println("ENCRYPTION_KEY=" + key)
	fmt.Println("\nIMPORTANT SECURITY NOTES:")
	fmt.Println("1. Keep this key secure! Anyone with this key can decrypt your data.")
	fmt.Println("2. Store a backup of this key in a secure location.")
	fmt.Println("3. If this key is lost, encrypted data cannot be recovered.")
}
