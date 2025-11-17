package aurelion

import (
	"github.com/anthanhphan/gosdk/transport/aurelion/validation"
)

// ValidateAndParsePublic combines BodyParser and Validate helper functions using public aurelion.Context.
func ValidateAndParsePublic(ctx Context, v interface{}) error {
	return validation.ValidateAndParse(ctx, v)
}
