package aurelion

import (
	"time"

	"github.com/anthanhphan/gosdk/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// requestResponseLoggingMiddleware creates middleware that logs request and response information.
//
// Input:
//   - logger: The logger instance to use for logging
//   - verbose: Whether to log verbose request/response data including body, query params, etc.
//
// Output:
//   - fiber.Handler: The middleware handler function
func requestResponseLoggingMiddleware(logger *zap.SugaredLogger, verbose bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := getRequestIDFromContext(c)
		log := logger.With(zap.String("request_id", requestID))

		// Log incoming request
		logFields := buildRequestLogFields(c, verbose)
		log.Infow("incoming request", logFields...)

		// Process request
		err := c.Next()

		// Log outgoing response
		duration := time.Since(start)
		responseFields := buildResponseLogFields(c, verbose, duration)
		log.Infow("request completed", responseFields...)

		return err
	}
}

// getRequestIDFromContext extracts request ID from fiber context.
// Returns "unknown" if request ID is not found or invalid type.
func getRequestIDFromContext(c *fiber.Ctx) string {
	requestID, ok := c.Locals(contextKeyRequestID).(string)
	if !ok {
		return "unknown"
	}
	return requestID
}

// buildRequestLogFields builds log fields for incoming request logging.
//
// Input:
//   - c: The fiber context containing request information
//   - verbose: Whether to include detailed information like body and query params
//
// Output:
//   - []interface{}: Slice of key-value pairs for structured logging
func buildRequestLogFields(c *fiber.Ctx, verbose bool) []interface{} {
	logFields := []interface{}{
		"method", c.Method(),
		"path", c.Path(),
		"ip", c.IP(),
		"user-agent", c.Get("User-Agent"),
	}

	if !verbose {
		return logFields
	}

	// Add query parameters if present
	if queryString := string(c.Request().URI().QueryString()); queryString != "" {
		logFields = append(logFields, "query", queryString)
	}

	// Add route params if present
	if params := c.AllParams(); len(params) > 0 {
		logFields = append(logFields, "params", params)
	}

	// Add formatted request body if present
	if body := string(c.Body()); body != "" {
		if formattedBody, err := utils.MarshalCompact(body); err == nil && formattedBody != "" {
			logFields = append(logFields, "body", formattedBody)
		} else {
			// If unmarshal fails, log original body as string
			logFields = append(logFields, "body", body)
		}
	}

	return logFields
}

// buildResponseLogFields builds log fields for outgoing response logging.
//
// Input:
//   - c: The fiber context containing response information
//   - verbose: Whether to include detailed information like response body
//   - duration: The request processing duration
//
// Output:
//   - []interface{}: Slice of key-value pairs for structured logging
func buildResponseLogFields(c *fiber.Ctx, verbose bool, duration time.Duration) []interface{} {
	responseFields := []interface{}{
		"method", c.Method(),
		"path", c.Path(),
		"http_code", c.Response().StatusCode(),
		"duration_ms", duration.Milliseconds(),
	}

	if !verbose {
		return responseFields
	}

	// Add formatted response body if present
	responseBody := string(c.Response().Body())
	if responseBody != "" {
		// Format response body to compact JSON
		formattedResponse, err := utils.MarshalCompact(responseBody)
		if err == nil && formattedResponse != "" {
			responseFields = append(responseFields, "response", formattedResponse)

			// Try to extract code from response
			if response, err := utils.Unmarshal(responseBody, APIResponse{}); err == nil {
				responseFields = append(responseFields, "code", response.Code)
			}
		}
	}

	return responseFields
}
