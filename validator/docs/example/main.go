// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/anthanhphan/gosdk/validator"
)

func main() {
	coreRules()
	comparisonRules()
	stringRules()
	formatRules()
	networkRules()
	collectionRules()
	nestedAndDive()
	instanceValidator()
	customRule()
	errorHandling()
}

func coreRules() {
	fmt.Println("-- Core: required, min, max, len, oneof --")

	type Product struct {
		Name     string `validate:"required,min=2,max=100"`
		SKU      string `validate:"required,len=8"`
		Category string `validate:"required,oneof=electronics clothing food"`
		Qty      int    `validate:"required,min=1,max=9999"`
	}

	check("valid", validator.Validate(Product{Name: "Laptop", SKU: "ELEC0001", Category: "electronics", Qty: 5}))
	check("invalid", validator.Validate(Product{Name: "X", SKU: "SHORT", Category: "toys", Qty: 0}))
	fmt.Println()
}

func comparisonRules() {
	fmt.Println("-- Comparison: gt, gte, lt, lte --")

	type Temp struct {
		Min      float64 `validate:"gte=-273.15"`
		Max      float64 `validate:"lte=1000"`
		Priority int     `validate:"required,gt=0,lt=100"`
	}

	check("valid", validator.Validate(Temp{Min: 20.5, Max: 99.9, Priority: 50}))
	check("invalid", validator.Validate(Temp{Min: -300, Max: 2000, Priority: 0}))
	fmt.Println()
}

func stringRules() {
	fmt.Println("-- String: contains, startswith, endswith, lowercase, uppercase, excludes --")

	type Doc struct {
		Title   string `validate:"required,startswith=DOC-"`
		Body    string `validate:"required,contains=approved"`
		Path    string `validate:"required,endswith=.pdf"`
		Code    string `validate:"required,uppercase"`
		Slug    string `validate:"required,lowercase"`
		Comment string `validate:"required,excludes=<script>"`
	}

	check("valid", validator.Validate(Doc{
		Title: "DOC-001", Body: "approved by board", Path: "/report.pdf",
		Code: "ABC", Slug: "hello-world", Comment: "ok",
	}))
	check("invalid", validator.Validate(Doc{
		Title: "report", Body: "pending", Path: "/report.docx",
		Code: "abc", Slug: "Hello", Comment: "<script>alert(1)",
	}))
	fmt.Println()
}

func formatRules() {
	fmt.Println("-- Format: email, url, numeric, alpha, alphanumeric, uuid, hexcolor, datetime --")

	type Form struct {
		Email    string `validate:"required,email"`
		Website  string `validate:"required,url"`
		Phone    string `validate:"required,numeric"`
		Name     string `validate:"required,alpha"`
		Username string `validate:"required,alphanumeric"`
		TraceID  string `validate:"required,uuid"`
		Color    string `validate:"required,hexcolor"`
		Date     string `validate:"required,datetime"`
	}

	check("valid", validator.Validate(Form{
		Email: "a@b.com", Website: "https://x.com", Phone: "123",
		Name: "John", Username: "john123",
		TraceID: "550e8400-e29b-41d4-a716-446655440000",
		Color: "#FF5733", Date: "2026-03-24T15:00:00Z",
	}))
	check("invalid", validator.Validate(Form{
		Email: "bad", Website: "ftp://x", Phone: "abc",
		Name: "J0hn", Username: "a@#",
		TraceID: "not-uuid", Color: "red", Date: "24/03/2026",
	}))

	// Custom datetime layout
	type Event struct {
		Start string `validate:"required,datetime=2006-01-02"`
	}
	check("custom layout", validator.Validate(Event{Start: "2026-03-24"}))
	check("bad layout", validator.Validate(Event{Start: "March 24"}))
	fmt.Println()
}

func networkRules() {
	fmt.Println("-- Network: ip, ipv4, ipv6 --")

	type Net struct {
		Any  string `validate:"required,ip"`
		V4   string `validate:"required,ipv4"`
		V6   string `validate:"required,ipv6"`
	}

	check("valid", validator.Validate(Net{Any: "10.0.0.1", V4: "10.0.0.1", V6: "::1"}))
	check("invalid", validator.Validate(Net{Any: "bad", V4: "::1", V6: "10.0.0.1"}))
	fmt.Println()
}

func collectionRules() {
	fmt.Println("-- Collection: notempty, unique --")

	type Survey struct {
		Answers []string          `validate:"notempty,unique"`
		Meta    map[string]string `validate:"notempty"`
	}

	check("valid", validator.Validate(Survey{Answers: []string{"A", "B"}, Meta: map[string]string{"v": "1"}}))
	check("empty", validator.Validate(Survey{}))
	check("duplicates", validator.Validate(Survey{Answers: []string{"A", "A"}, Meta: map[string]string{"v": "1"}}))
	fmt.Println()
}

func nestedAndDive() {
	fmt.Println("-- Nested Structs + Dive --")

	type Addr struct {
		City string `validate:"required"`
		Zip  string `validate:"required,numeric,len=5"`
	}
	type User struct {
		Name  string   `validate:"required"`
		Addr  *Addr    // auto-validated, nil ptr skipped
		Ports []int    `validate:"notempty,dive,min=1,max=65535"`
		Tags  []string `validate:"dive,required,min=2"`
	}

	check("valid", validator.Validate(User{
		Name: "Alice", Addr: &Addr{City: "NYC", Zip: "10001"},
		Ports: []int{80, 443}, Tags: []string{"go", "api"},
	}))
	check("nil ptr", validator.Validate(User{Name: "Bob", Ports: []int{80}}))
	check("invalid nested+dive", validator.Validate(User{
		Name: "", Addr: &Addr{City: "", Zip: "ABC"},
		Ports: []int{0, 70000}, Tags: []string{"a", ""},
	}))
	fmt.Println()
}

func instanceValidator() {
	fmt.Println("-- Instance: WithFieldNameTag, WithStopOnFirstError --")

	type Login struct {
		User string `json:"username" validate:"required,min=3"`
		Pass string `json:"password" validate:"required,min=8"`
	}

	v := validator.New(validator.WithFieldNameTag("json"))
	check("json tags", v.ValidateStruct(Login{User: "ab"}))

	v2 := validator.New(validator.WithStopOnFirstError(true))
	check("stop on first", v2.ValidateStruct(Login{}))
	fmt.Println()
}

func customRule() {
	fmt.Println("-- Custom Rules --")

	validator.RegisterValidationRule("even", func(f string, v reflect.Value, _ string) *validator.ValidationError {
		if v.Kind() == reflect.Int && v.Int()%2 != 0 {
			return &validator.ValidationError{Field: f, Message: "must be even"}
		}
		return nil
	})
	validator.RegisterValidationRule("nospace", func(f string, v reflect.Value, _ string) *validator.ValidationError {
		if v.Kind() == reflect.String && strings.Contains(v.String(), " ") {
			return &validator.ValidationError{Field: f, Message: "must not contain spaces"}
		}
		return nil
	})

	type T struct {
		N    int    `validate:"required,even"`
		Slug string `validate:"required,nospace"`
	}
	check("valid", validator.Validate(T{N: 4, Slug: "ok"}))
	check("invalid", validator.Validate(T{N: 3, Slug: "has space"}))
	fmt.Println()
}

func errorHandling() {
	fmt.Println("-- Error Handling --")

	type Req struct {
		Name string `json:"name" validate:"required,min=3"`
	}

	v := validator.New(validator.WithFieldNameTag("json"))
	err := v.ValidateStruct(Req{Name: "ab"})
	if err == nil {
		return
	}
	errs := err.(validator.ValidationErrors)

	fmt.Println("  .Error():", errs.Error())
	fmt.Println("  .ToArray():", errs.ToArray())
	for _, e := range errs {
		fmt.Printf("  field=%s message=%s\n", e.Field, e.Message)
	}
}

func check(label string, err error) {
	if err == nil {
		fmt.Printf("  %s: OK\n", label)
		return
	}
	fmt.Printf("  %s:\n", label)
	for _, e := range err.(validator.ValidationErrors) {
		fmt.Printf("    %-20s %s\n", e.Field, e.Message)
	}
}
