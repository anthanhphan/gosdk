// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"fmt"
	"reflect"

	"github.com/anthanhphan/gosdk/validator"
)

// ============================================================================
// Example structs
// ============================================================================

type Address struct {
	Street string `json:"street" validate:"required,min=5"`
	City   string `json:"city" validate:"required"`
	Zip    string `json:"zip" validate:"required,numeric,len=5"`
}

type User struct {
	Name     string   `json:"name" validate:"required,min=2,max=50,alpha"`
	Email    string   `json:"email" validate:"required,email"`
	Age      int      `json:"age" validate:"required,gte=18,lte=120"`
	Role     string   `json:"role" validate:"required,oneof=admin user guest"`
	Website  string   `json:"website" validate:"url"`
	Tags     []string `json:"tags" validate:"notempty,unique,dive,required,min=1"`
	Address  Address  `json:"address"`
	password string
}

type ServerConfig struct {
	Host     string `validate:"required,ip"`
	Port     int    `validate:"required,min=1,max=65535"`
	Protocol string `validate:"required,oneof=http https grpc"`
	LogLevel string `validate:"required,lowercase"`
}

func main() {
	fmt.Println("=== Validator Example ===")
	fmt.Println()

	// 1. Basic validation using package-level Validate
	basicValidation()

	// 2. Instance-based validation with JSON field names
	instanceValidation()

	// 3. Custom rule registration
	customRuleExample()

	// 4. Server config validation
	configValidation()
}

func basicValidation() {
	fmt.Println("--- Basic Validation ---")

	user := User{
		Name:  "J", // too short
		Email: "not-an-email",
		Age:   15, // under 18
		Role:  "superadmin",
		Tags:  []string{"go", ""},
		Address: Address{
			Street: "Hi", // too short
			City:   "",   // empty
			Zip:    "ABC",
		},
	}

	err := validator.Validate(user)
	if err == nil {
		fmt.Println("  No errors (unexpected!)")
		return
	}

	errs := err.(validator.ValidationErrors)
	for _, e := range errs {
		fmt.Printf("  %-20s %s\n", e.Field, e.Message)
	}

	fmt.Println()
	fmt.Println("  ToArray output:")
	for _, item := range errs.ToArray() {
		fmt.Printf("    field=%-20s message=%s\n", item["field"], item["message"])
	}
	fmt.Println()
}

func instanceValidation() {
	fmt.Println("--- Instance with JSON Tags ---")

	v := validator.New(
		validator.WithFieldNameTag("json"),
		validator.WithStopOnFirstError(true),
	)

	user := User{Name: "", Email: ""}
	err := v.ValidateStruct(user)
	if err == nil {
		fmt.Println("  No errors (unexpected!)")
		return
	}

	errs := err.(validator.ValidationErrors)
	fmt.Printf("  Stopped on first error: %s\n", errs[0].Error())
	fmt.Println()
}

func customRuleExample() {
	fmt.Println("--- Custom Rule ---")

	// Register a custom "even" rule
	validator.RegisterValidationRule("even", func(field string, value reflect.Value, _ string) *validator.ValidationError {
		if value.Kind() == reflect.Int && value.Int()%2 != 0 {
			return &validator.ValidationError{Field: field, Message: "must be an even number"}
		}
		return nil
	})

	type Pair struct {
		Count int `validate:"required,even"`
	}

	// Valid
	if err := validator.Validate(Pair{Count: 4}); err != nil {
		fmt.Printf("  Count=4: %v\n", err)
	} else {
		fmt.Println("  Count=4: OK")
	}

	// Invalid
	if err := validator.Validate(Pair{Count: 7}); err != nil {
		fmt.Printf("  Count=7: %v\n", err)
	}
	fmt.Println()
}

func configValidation() {
	fmt.Println("--- Config Validation ---")

	// Valid config
	good := ServerConfig{
		Host:     "192.168.1.1",
		Port:     8080,
		Protocol: "https",
		LogLevel: "debug",
	}
	if err := validator.Validate(good); err != nil {
		fmt.Printf("  Valid config failed: %v\n", err)
	} else {
		fmt.Println("  Valid config: OK")
	}

	// Invalid config
	bad := ServerConfig{
		Host:     "not-an-ip",
		Port:     99999,
		Protocol: "ftp",
		LogLevel: "DEBUG",
	}
	if err := validator.Validate(bad); err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			fmt.Printf("  %-12s %s\n", e.Field, e.Message)
		}
	}
	fmt.Println()
}
