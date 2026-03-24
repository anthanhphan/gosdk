# Validator Package

A high-performance, zero-allocation (after first call) struct validation engine for Go. Uses struct tags to define rules — all parameters are pre-parsed and cached on first use.

## Installation

```bash
go get github.com/anthanhphan/gosdk/validator
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/anthanhphan/gosdk/validator"
)

type User struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
	Age   int    `validate:"required,gte=18,lte=120"`
	Role  string `validate:"required,oneof=admin user guest"`
}

func main() {
	user := User{Name: "A", Email: "bad", Age: 15, Role: "hacker"}
	if err := validator.Validate(user); err != nil {
		fmt.Println(err)
		// Name: must be at least 2 characters; Email: must be a valid email address;
		// Age: must be greater than or equal to 18; Role: must be one of: admin user guest
	}
}
```

## API

### `Validate(v any) error`

Package-level function. Validates a struct using `validate` struct tags.

```go
err := validator.Validate(myStruct)
```

### `New(opts ...ValidatorOption) *Validator`

Creates a configurable validator instance.

```go
v := validator.New(
    validator.WithFieldNameTag("json"),       // use json tag names in errors
    validator.WithStopOnFirstError(true),     // stop after first error
)
err := v.ValidateStruct(myStruct)
```

### `RegisterValidationRule(name string, rule CustomValidationRule)`

Registers a custom validation rule globally.

```go
validator.RegisterValidationRule("even", func(field string, value reflect.Value, param string) *validator.ValidationError {
    if value.Kind() == reflect.Int && value.Int()%2 != 0 {
        return &validator.ValidationError{Field: field, Message: "must be even"}
    }
    return nil
})
```

## Available Rules

### Core

| Rule | Tag | Description |
|------|-----|-------------|
| Required | `required` | Field must not be zero value |
| Min | `min=N` | Min length (string/slice) or min value (number) |
| Max | `max=N` | Max length (string/slice) or max value (number) |
| Len | `len=N` | Exact length (string/slice/map) |
| OneOf | `oneof=a b c` | Value must be one of the listed values |

### Comparison

| Rule | Tag | Description |
|------|-----|-------------|
| Greater Than | `gt=N` | Must be > N |
| Greater Than or Equal | `gte=N` | Must be >= N |
| Less Than | `lt=N` | Must be < N |
| Less Than or Equal | `lte=N` | Must be <= N  |

### String

| Rule | Tag | Description |
|------|-----|-------------|
| Contains | `contains=substr` | Must contain substring |
| Starts With | `startswith=prefix` | Must start with prefix |
| Ends With | `endswith=suffix` | Must end with suffix |
| Lowercase | `lowercase` | Must be lowercase |
| Uppercase | `uppercase` | Must be uppercase |
| Excludes | `excludes=substr` | Must not contain substring |

### Format

| Rule | Tag | Description |
|------|-----|-------------|
| Email | `email` | Valid email address |
| URL | `url` | Valid HTTP/HTTPS URL |
| Numeric | `numeric` | Digits only |
| Alpha | `alpha` | Letters only |
| Alphanumeric | `alphanumeric` | Letters and digits only |
| UUID | `uuid` | Valid UUID format |
| Hex Color | `hexcolor` | Valid hex color (#FFF or #FFFFFF) |
| Datetime | `datetime` or `datetime=layout` | Valid datetime (default: RFC3339) |

### Network

| Rule | Tag | Description |
|------|-----|-------------|
| IP | `ip` | Valid IP address (v4 or v6) |
| IPv4 | `ipv4` | Valid IPv4 address |
| IPv6 | `ipv6` | Valid IPv6 address |

### Collection

| Rule | Tag | Description |
|------|-----|-------------|
| Not Empty | `notempty` | Slice/map/string must not be empty |
| Unique | `unique` | Slice must contain unique values |
| Dive | `dive` | Validate each element in slice/map |

## Advanced Usage

### Combining Rules

```go
type Config struct {
    Host string `validate:"required,ip"`
    Port int    `validate:"required,min=1,max=65535"`
}
```

### Dive (Element Validation)

```go
type Request struct {
    Tags []string `validate:"notempty,dive,required,min=1"`
}
```

Rules before `dive` apply to the field itself. Rules after `dive` apply to each element.

### Nested Structs

```go
type Address struct {
    City string `validate:"required"`
}
type User struct {
    Name    string  `validate:"required"`
    Address Address // automatically validated
}
```

Nested struct errors are prefixed: `Address.City: is required`

### JSON Field Names in Errors

```go
v := validator.New(validator.WithFieldNameTag("json"))
err := v.ValidateStruct(myStruct)
// errors use json tag names: "first_name: is required"
```

### Error Handling

```go
err := validator.Validate(myStruct)
if err != nil {
    if errs, ok := err.(validator.ValidationErrors); ok {
        // Structured errors
        for _, e := range errs {
            fmt.Printf("field=%s msg=%s\n", e.Field, e.Message)
        }
        // For API responses
        arr := errs.ToArray() // []map[string]string{{"field":"name","message":"is required"}}
    }
}
```

## Performance

After the first validation call for each struct type, all subsequent calls execute with:

- **Zero string parsing** — all tag rules pre-parsed at cache time
- **Zero map lookups** — rule handlers pre-resolved to function pointers
- **Zero regex recompilation** — patterns compiled once at package init
- **Zero allocation** for field metadata — cached per struct type
