package router

// Types defined locally to avoid import cycles with main aurelion package.

// Method represents HTTP methods.
type Method string

// Handler defines the function signature for route handlers.
type Handler func(ContextInterface) error

// Middleware defines the function signature for middleware functions.
type Middleware func(ContextInterface) error

// ContextInterface defines the minimal interface needed by router functions.
type ContextInterface interface {
	Status(status int) ContextInterface
	JSON(data interface{}) error
	Locals(key string, value ...interface{}) interface{}
	GetAllLocals() map[string]interface{}
	Next() error
	Method() string
	Path() string
	Params(key string, defaultValue ...string) string
	AllParams() map[string]string
	ParamsParser(out interface{}) error
	Query(key string, defaultValue ...string) string
	AllQueries() map[string]string
	QueryParser(out interface{}) error
	Body() []byte
	BodyParser(out interface{}) error
}

// CORSConfig represents CORS configuration (duplicated to avoid import cycle).
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// HTTP method constants.
const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH"
	MethodDelete  Method = "DELETE"
	MethodHead    Method = "HEAD"
	MethodOptions Method = "OPTIONS"
)
