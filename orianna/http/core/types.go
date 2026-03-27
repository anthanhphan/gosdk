// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "time"

// Handler defines the function signature for route handlers.
// Handlers are terminal — they process the request and send a response.
// They must NOT call ctx.Next(), as they are the end of the middleware chain.
type Handler func(Context) error

// Middleware defines the function signature for middleware functions.
// Middleware must call ctx.Next() to pass control to the next handler in the chain.
// Failing to call ctx.Next() short-circuits the chain (useful for auth, rate limiting, etc.).
//
// NOTE: Handler and Middleware share the same underlying type signature.
// The distinction is semantic: middleware chains, handlers terminate.
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

// methodNames maps Method values to their string representations.
var methodNames = [...]string{
	GET:     "GET",
	POST:    "POST",
	PUT:     "PUT",
	PATCH:   "PATCH",
	DELETE:  "DELETE",
	HEAD:    "HEAD",
	OPTIONS: "OPTIONS",
}

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
	if int(m) >= 0 && int(m) < len(methodNames) {
		return methodNames[m]
	}
	return "UNKNOWN"
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
