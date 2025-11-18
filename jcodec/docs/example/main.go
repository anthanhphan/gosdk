// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anthanhphan/gosdk/jcodec"
)

// User represents a user in the system.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

// Config represents application configuration.
type Config struct {
	DatabaseURL string            `json:"database_url"`
	Port        int               `json:"port"`
	Debug       bool              `json:"debug"`
	Headers     map[string]string `json:"headers"`
}

func main() {
	fmt.Println("=== jcodec Example ===")
	fmt.Println()

	// Example 1: Basic Marshal and Unmarshal
	exampleBasic()

	// Example 2: Marshal and Unmarshal with complex structures
	exampleComplex()

	// Example 3: Round-trip marshaling
	exampleRoundTrip()
}

// exampleBasic demonstrates basic Marshal and Unmarshal usage.
func exampleBasic() {
	fmt.Println("1. Basic Marshal and Unmarshal:")

	user := User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		Active:    true,
	}

	// Marshal
	data, err := jcodec.Marshal(user)
	if err != nil {
		log.Fatal("Marshal error:", err)
	}
	fmt.Printf("   Marshaled: %s\n", string(data))

	// Unmarshal
	var unmarshaled User
	err = jcodec.Unmarshal(data, &unmarshaled)
	if err != nil {
		log.Fatal("Unmarshal error:", err)
	}
	fmt.Printf("   Unmarshaled: %+v\n\n", unmarshaled)
}

// exampleComplex demonstrates Marshal and Unmarshal with complex structures.
func exampleComplex() {
	fmt.Println("2. Complex Structures:")

	config := Config{
		DatabaseURL: "postgres://localhost:5432/mydb",
		Port:        8080,
		Debug:       true,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token",
		},
	}

	// Marshal
	data, err := jcodec.Marshal(config)
	if err != nil {
		log.Fatal("Marshal error:", err)
	}
	fmt.Printf("   Marshaled: %s\n", string(data))

	// Unmarshal
	var unmarshaled Config
	err = jcodec.Unmarshal(data, &unmarshaled)
	if err != nil {
		log.Fatal("Unmarshal error:", err)
	}
	fmt.Printf("   Unmarshaled: %+v\n\n", unmarshaled)
}

// exampleRoundTrip demonstrates round-trip marshaling.
func exampleRoundTrip() {
	fmt.Println("3. Round-Trip Marshaling:")
	fmt.Println("   Marshal -> Unmarshal -> Verify")

	original := User{
		ID:        4,
		Name:      "Alice Williams",
		Email:     "alice@example.com",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Active:    true,
	}

	// Marshal
	data, err := jcodec.Marshal(original)
	if err != nil {
		log.Fatal("Marshal error:", err)
	}

	// Unmarshal
	var roundTrip User
	err = jcodec.Unmarshal(data, &roundTrip)
	if err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	// Verify
	if original.ID != roundTrip.ID {
		log.Fatalf("ID mismatch: %d != %d", original.ID, roundTrip.ID)
	}
	if original.Name != roundTrip.Name {
		log.Fatalf("Name mismatch: %s != %s", original.Name, roundTrip.Name)
	}
	if original.Email != roundTrip.Email {
		log.Fatalf("Email mismatch: %s != %s", original.Email, roundTrip.Email)
	}

	fmt.Printf("   Original:  %+v\n", original)
	fmt.Printf("   RoundTrip: %+v\n", roundTrip)
	fmt.Println("   ✓ Round-trip successful!")
}
