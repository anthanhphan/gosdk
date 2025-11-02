package aurelion

// Map is a convenient alias for map[string]interface{} used for flexible data structures
type Map map[string]interface{}

// Handler defines the function signature for route handlers.
// Handlers receive a Context and return an error.
// The error can be used for control flow or logged by middleware.
type Handler func(Context) error

// Middleware defines the function signature for middleware functions.
// Middleware can pre-process requests, post-process responses,
// or short-circuit the request handling by not calling ctx.Next().
type Middleware func(Context) error

// Method represents HTTP methods as type-safe constants
type Method int

// HTTP method constants
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
//   - string: The HTTP method name (GET, POST, etc.), or "UNKNOWN" if invalid
//
// Example:
//
//	method := aurelion.GET
//	fmt.Println(method.String()) // Output: "GET"
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
