package middleware

import (
	"time"

	"github.com/anthanhphan/gosdk/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequestResponseLogger creates middleware that logs request and response information.
func RequestResponseLogger(logger *zap.SugaredLogger, verbose bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID, _ := c.Locals(ContextKeyRequestID).(string)
		log := logger.With(zap.String("request_id", requestID))

		logFields := buildRequestLogFields(c, verbose)
		log.Infow("incoming request", logFields...)

		err := c.Next()

		duration := time.Since(start)
		responseFields := buildResponseLogFields(c, verbose, duration)
		log.Infow("request completed", responseFields...)

		return err
	}
}

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

	if queryString := string(c.Request().URI().QueryString()); queryString != "" {
		logFields = append(logFields, "query", queryString)
	}

	if params := c.AllParams(); len(params) > 0 {
		logFields = append(logFields, "params", params)
	}

	if body := string(c.Body()); body != "" {
		if formattedBody, err := utils.MarshalCompact(body); err == nil && formattedBody != "" {
			logFields = append(logFields, "body", formattedBody)
		} else {
			logFields = append(logFields, "body", body)
		}
	}

	return logFields
}

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

	responseBody := string(c.Response().Body())
	if responseBody != "" {
		formattedResponse, err := utils.MarshalCompact(responseBody)
		if err == nil && formattedResponse != "" {
			responseFields = append(responseFields, "response", formattedResponse)

			// Try to parse response body as JSON to extract code field
			// Using interface{} to avoid importing response package
			var respData map[string]interface{}
			if _, err := utils.Unmarshal(responseBody, &respData); err == nil {
				if code, ok := respData["code"]; ok {
					responseFields = append(responseFields, "code", code)
				}
			}
		}
	}

	return responseFields
}
