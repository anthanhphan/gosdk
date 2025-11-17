package validation

// ContextInterface defines the minimal interface needed by validation functions.
type ContextInterface interface {
	BodyParser(out interface{}) error
}
