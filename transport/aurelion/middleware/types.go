package middleware

// ContextInterface defines the minimal interface needed by middleware functions.
// This avoids importing the main aurelion package and breaking import cycles.
type ContextInterface interface {
	Locals(key string, value ...interface{}) interface{}
	GetAllLocals() map[string]interface{}
	Next() error
}

// MiddlewareFunc defines the middleware function signature.
type MiddlewareFunc func(ContextInterface) error
