package main

import (
	"fmt"
	"log"

	"auth0-server/internal/infrastructure/crypto"
)

func main() {
	fmt.Println("=== Password Hashing Test ===")

	// Create password hasher
	hasher := crypto.DefaultPasswordHasher()

	// Test password
	plaintext := "mySecurePassword123"
	fmt.Printf("Original password: %s\n", plaintext)

	// Hash the password
	hashed, err := hasher.Hash(plaintext)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	fmt.Printf("Hashed password: %s\n", hashed)
	fmt.Printf("Hash length: %d\n", len(hashed))
	fmt.Printf("Is bcrypt format: %t\n", len(hashed) == 60 && hashed[:4] == "$2a$" || hashed[:4] == "$2b$" || hashed[:4] == "$2y$")

	// Test correct password
	fmt.Println("\nTesting correct password validation:")
	err = hasher.Compare(hashed, plaintext)
	valid := err == nil
	if err != nil && err.Error() != "password does not match" {
		log.Fatalf("Failed to validate password: %v", err)
	}
	fmt.Printf("Password validation result: %t\n", valid)

	// Test incorrect password
	fmt.Println("\nTesting incorrect password validation:")
	err = hasher.Compare(hashed, "wrongPassword")
	invalidPasswordRejected := err != nil // Should return error for wrong password
	if err != nil && err.Error() != "password does not match" {
		log.Fatalf("Failed to validate password: %v", err)
	}
	fmt.Printf("Wrong password validation result: %t\n", !invalidPasswordRejected)

	fmt.Println("\n=== Test Results ===")
	fmt.Printf("✅ Password hashing: %s\n", getStatus(err == nil))
	fmt.Printf("✅ Correct password validation: %s\n", getStatus(valid))
	fmt.Printf("✅ Incorrect password rejection: %s\n", getStatus(invalidPasswordRejected))
}

func getStatus(success bool) string {
	if success {
		return "PASS"
	}
	return "FAIL"
}
