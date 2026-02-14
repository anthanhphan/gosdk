// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "time"

// Handler defines the function signature for route handlers.
type Handler func(Context) error

// Middleware defines the function signature for middleware functions.
type Middleware func(Context) error

// Method represents HTTP methods as type-safe constants.
type Method int

const (
	GET Method = iota
	POST
	PUT
	PATCH
	DELETE
	HEAD
	OPTIONS
)

// String returns the string representation of the HTTP method.
//
// Output:
//   - string: The HTTP method name
//
// Example:
//
//	method := core.GET
//	fmt.Println(method.String()) // "GET"
func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case PATCH:
		return "PATCH"
	case DELETE:
		return "DELETE"
	case HEAD:
		return "HEAD"
	case OPTIONS:
		return "OPTIONS"
	default:
		return "UNKNOWN"
	}
}

// Cookie represents an HTTP cookie.
type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	Expires  time.Time
	Secure   bool
	HTTPOnly bool
	SameSite string
}
