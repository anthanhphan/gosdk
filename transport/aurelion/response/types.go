package response

// ContextInterface defines the minimal interface needed by response functions.
// This avoids importing the main aurelion package and breaking import cycles.
type ContextInterface interface {
	Status(status int) ContextInterface
	JSON(data interface{}) error
	Locals(key string, value ...interface{}) interface{}
	Next() error
	GetAllLocals() map[string]interface{}
}

// Map is a convenient alias for map[string]interface{}.
type Map map[string]interface{}
