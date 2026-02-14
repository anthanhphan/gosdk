// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"bytes"
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

	// Example 4: Streaming API
	exampleStreaming()

	// Example 5: RawMessage
	exampleRawMessage()

	// Example 6: Utility Functions
	exampleUtilityFunctions()
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
	fmt.Println("   - Round-trip successful!")
	fmt.Println()
}

// exampleStreaming demonstrates the Encoder and Decoder API.
func exampleStreaming() {
	fmt.Println("4. Streaming API:")

	// Encoder
	fmt.Println("   Encoder:")
	user := User{ID: 5, Name: "Stream User", Email: "stream@example.com"}

	// In a real app, you might write to a file or network connection
	// Here we write to a buffer for demonstration
	var buf bytes.Buffer
	enc := jcodec.NewEncoder(&buf)
	enc.SetIndent("", "  ")

	if err := enc.Encode(user); err != nil {
		log.Fatal("Encode error:", err)
	}
	fmt.Printf("   Encoded:\n%s\n", buf.String())

	// Decoder
	fmt.Println("   Decoder:")
	dec := jcodec.NewDecoder(&buf)
	var decoded User
	if err := dec.Decode(&decoded); err != nil {
		log.Fatal("Decode error:", err)
	}
	fmt.Printf("   Decoded: %+v\n\n", decoded)
}

// exampleRawMessage demonstrates using RawMessage for delayed decoding.
func exampleRawMessage() {
	fmt.Println("5. RawMessage:")

	type Event struct {
		Type    string            `json:"type"`
		Payload jcodec.RawMessage `json:"payload"`
	}

	jsonData := []byte(`{
		"type": "user_created",
		"payload": {"id": 10, "name": "Raw User"}
	}`)

	var event Event
	if err := jcodec.Unmarshal(jsonData, &event); err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	fmt.Printf("   Event Type: %s\n", event.Type)
	fmt.Printf("   Raw Payload: %s\n", string(event.Payload))

	if event.Type == "user_created" {
		var user User
		if err := jcodec.Unmarshal(event.Payload, &user); err != nil {
			log.Fatal("Unmarshal payload error:", err)
		}
		fmt.Printf("   Parsed Payload: %+v\n\n", user)
	}
}

// exampleUtilityFunctions demonstrates helper functions like Indent and Compact.
func exampleUtilityFunctions() {
	fmt.Println("6. Utility Functions:")

	src := []byte(`{"id":1,"name":"Compact User"}`)

	// Indent
	var indentBuf bytes.Buffer
	if err := jcodec.Indent(&indentBuf, src, "", "    "); err != nil {
		log.Fatal("Indent error:", err)
	}
	fmt.Printf("   Indented:\n%s\n", indentBuf.String())

	// Compact
	var compactBuf bytes.Buffer
	if err := jcodec.Compact(&compactBuf, indentBuf.Bytes()); err != nil {
		log.Fatal("Compact error:", err)
	}
	fmt.Printf("   Compacted: %s\n", compactBuf.String())
}
