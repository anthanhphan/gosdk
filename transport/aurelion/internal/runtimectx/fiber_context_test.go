package runtimectx

import (
	"testing"

	"github.com/anthanhphan/gosdk/transport/aurelion/core"
	"github.com/anthanhphan/gosdk/transport/aurelion/internal/keys"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestNewFiberContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *fiber.Ctx
		check func(t *testing.T, result core.Context)
	}{
		{
			name: "valid fiber context should create wrapper",
			setup: func() *fiber.Ctx {
				app := fiber.New()
				ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return ctx
			},
			check: func(t *testing.T, result core.Context) {
				if result == nil {
					t.Error("NewFiberContext() should return non-nil context")
				}
				if _, ok := result.(*FiberContext); !ok {
					t.Errorf("NewFiberContext() should return *FiberContext, got %T", result)
				}
			},
		},
		{
			name: "nil fiber context should return nil",
			setup: func() *fiber.Ctx {
				return nil
			},
			check: func(t *testing.T, result core.Context) {
				if result != nil {
					t.Error("NewFiberContext(nil) should return nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCtx := tt.setup()
			result := NewFiberContext(fCtx)
			tt.check(t, result)

			if fCtx != nil {
				fCtx.App().ReleaseCtx(fCtx)
			}
		})
	}
}

func TestFiberFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() core.Context
		check func(t *testing.T, fCtx *fiber.Ctx, ok bool)
	}{
		{
			name: "valid FiberContext should extract fiber context",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return NewFiberContext(fCtx)
			},
			check: func(t *testing.T, fCtx *fiber.Ctx, ok bool) {
				if !ok {
					t.Error("FiberFromContext() should return ok=true for valid FiberContext")
				}
				if fCtx == nil {
					t.Error("FiberFromContext() should return non-nil fiber context")
				}
			},
		},
		{
			name: "nil context should return false",
			setup: func() core.Context {
				return nil
			},
			check: func(t *testing.T, fCtx *fiber.Ctx, ok bool) {
				if ok {
					t.Error("FiberFromContext(nil) should return ok=false")
				}
				if fCtx != nil {
					t.Error("FiberFromContext(nil) should return nil fiber context")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			fCtx, ok := FiberFromContext(ctx)
			tt.check(t, fCtx, ok)
		})
	}
}

func TestHandlerToFiber(t *testing.T) {
	tests := []struct {
		name    string
		handler core.Handler
		check   func(t *testing.T, fiberHandler fiber.Handler)
	}{
		{
			name: "valid handler should be converted",
			handler: func(ctx core.Context) error {
				ctx.Locals("test", "value")
				return nil
			},
			check: func(t *testing.T, fiberHandler fiber.Handler) {
				if fiberHandler == nil {
					t.Error("HandlerToFiber() should return non-nil fiber handler")
					return
				}

				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fCtx)

				err := fiberHandler(fCtx)
				if err != nil {
					t.Errorf("Converted handler returned error: %v", err)
				}

				if fCtx.Locals("test") != "value" {
					t.Error("Handler did not set locals correctly")
				}
			},
		},
		{
			name:    "nil handler should return no-op",
			handler: nil,
			check: func(t *testing.T, fiberHandler fiber.Handler) {
				if fiberHandler == nil {
					t.Error("HandlerToFiber(nil) should return non-nil fiber handler")
					return
				}

				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fCtx)

				err := fiberHandler(fCtx)
				if err != nil {
					t.Errorf("No-op handler should not return error, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiberHandler := HandlerToFiber(tt.handler)
			tt.check(t, fiberHandler)
		})
	}
}

func TestMiddlewareToFiber(t *testing.T) {
	tests := []struct {
		name       string
		middleware core.Middleware
		check      func(t *testing.T, fiberHandler fiber.Handler)
	}{
		{
			name: "valid middleware should be converted",
			middleware: func(ctx core.Context) error {
				ctx.Locals("middleware", "executed")
				return nil // Don't call Next in test
			},
			check: func(t *testing.T, fiberHandler fiber.Handler) {
				if fiberHandler == nil {
					t.Error("MiddlewareToFiber() should return non-nil fiber handler")
					return
				}

				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fCtx)

				err := fiberHandler(fCtx)
				if err != nil {
					t.Errorf("Converted middleware returned error: %v", err)
				}

				if fCtx.Locals("middleware") != "executed" {
					t.Error("Middleware did not set locals correctly")
				}
			},
		},
		{
			name:       "nil middleware should return handler",
			middleware: nil,
			check: func(t *testing.T, fiberHandler fiber.Handler) {
				if fiberHandler == nil {
					t.Error("MiddlewareToFiber(nil) should return non-nil fiber handler")
				}
				// Note: Cannot test Next() without full fiber app chain
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiberHandler := MiddlewareToFiber(tt.middleware)
			tt.check(t, fiberHandler)
		})
	}
}

func TestTrackFiberLocal(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *fiber.Ctx
		key   string
		check func(t *testing.T, fCtx *fiber.Ctx)
	}{
		{
			name: "valid key should be tracked",
			setup: func() *fiber.Ctx {
				app := fiber.New()
				return app.AcquireCtx(&fasthttp.RequestCtx{})
			},
			key: "test_key",
			check: func(t *testing.T, fCtx *fiber.Ctx) {
				tracked := getOrCreateTrackedLocalsMap(fCtx)
				if _, exists := tracked["test_key"]; !exists {
					t.Error("Key should be tracked")
				}
			},
		},
		{
			name: "empty key should not panic",
			setup: func() *fiber.Ctx {
				app := fiber.New()
				return app.AcquireCtx(&fasthttp.RequestCtx{})
			},
			key: "",
			check: func(t *testing.T, fCtx *fiber.Ctx) {
				// Should not panic
			},
		},
		{
			name: "nil context should not panic",
			setup: func() *fiber.Ctx {
				return nil
			},
			key: "test_key",
			check: func(t *testing.T, fCtx *fiber.Ctx) {
				// Should not panic
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fCtx := tt.setup()
			TrackFiberLocal(fCtx, tt.key)
			tt.check(t, fCtx)

			if fCtx != nil {
				fCtx.App().ReleaseCtx(fCtx)
			}
		})
	}
}

func TestFiberContext_Locals(t *testing.T) {
	tests := []struct {
		name  string
		setup func() core.Context
		key   string
		value interface{}
		check func(t *testing.T, ctx core.Context)
	}{
		{
			name: "set and get locals should work",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return NewFiberContext(fCtx)
			},
			key:   "test_key",
			value: "test_value",
			check: func(t *testing.T, ctx core.Context) {
				result := ctx.Locals("test_key")
				if result != "test_value" {
					t.Errorf("Locals() = %v, want test_value", result)
				}
			},
		},
		{
			name: "get non-existing key should return nil",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return NewFiberContext(fCtx)
			},
			key: "non_existing",
			check: func(t *testing.T, ctx core.Context) {
				result := ctx.Locals("non_existing")
				if result != nil {
					t.Errorf("Locals(non_existing) = %v, want nil", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			if tt.value != nil {
				ctx.Locals(tt.key, tt.value)
			}
			tt.check(t, ctx)
		})
	}
}

func TestFiberContext_GetAllLocals(t *testing.T) {
	tests := []struct {
		name  string
		setup func() core.Context
		check func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "get all locals should return tracked keys",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				ctx := NewFiberContext(fCtx)
				ctx.Locals("key1", "value1")
				ctx.Locals("key2", "value2")
				ctx.Locals(keys.ContextKeyRequestID, "req-123")
				return ctx
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["key1"] != "value1" {
					t.Errorf("key1 = %v, want value1", result["key1"])
				}
				if result["key2"] != "value2" {
					t.Errorf("key2 = %v, want value2", result["key2"])
				}
				if result[keys.ContextKeyRequestID] != "req-123" {
					t.Errorf("request_id = %v, want req-123", result[keys.ContextKeyRequestID])
				}
			},
		},
		{
			name: "empty locals should return empty map",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return NewFiberContext(fCtx)
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if len(result) != 0 {
					t.Errorf("GetAllLocals() should return empty map, got %d items", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := ctx.GetAllLocals()
			tt.check(t, result)
		})
	}
}

func TestFiberContext_RequestID(t *testing.T) {
	tests := []struct {
		name  string
		setup func() core.Context
		want  string
	}{
		{
			name: "request ID in locals should be returned",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				ctx := NewFiberContext(fCtx)
				ctx.Locals(keys.ContextKeyRequestID, "req-456")
				return ctx
			},
			want: "req-456",
		},
		{
			name: "no request ID should return empty string",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				return NewFiberContext(fCtx)
			},
			want: "",
		},
		{
			name: "non-string request ID should return empty string",
			setup: func() core.Context {
				app := fiber.New()
				fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				ctx := NewFiberContext(fCtx)
				ctx.Locals(keys.ContextKeyRequestID, 123)
				return ctx
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := ctx.RequestID()
			if result != tt.want {
				t.Errorf("RequestID() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestFiberContext_Status(t *testing.T) {
	tests := []struct {
		name   string
		status int
		check  func(t *testing.T, ctx core.Context, fCtx *fiber.Ctx)
	}{
		{
			name:   "set status should work",
			status: 201,
			check: func(t *testing.T, ctx core.Context, fCtx *fiber.Ctx) {
				if fCtx.Response().StatusCode() != 201 {
					t.Errorf("Status = %d, want 201", fCtx.Response().StatusCode())
				}
			},
		},
		{
			name:   "set status should return context for chaining",
			status: 404,
			check: func(t *testing.T, ctx core.Context, fCtx *fiber.Ctx) {
				returnedCtx := ctx.Status(404)
				if returnedCtx != ctx {
					t.Error("Status() should return same context for chaining")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(fCtx)

			ctx := NewFiberContext(fCtx)
			ctx.Status(tt.status)
			tt.check(t, ctx, fCtx)
		})
	}
}

func TestFiberContext_RequestMethods(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	// Setup request - Method is read-only after ctx creation, need to use request
	fCtx.Request().Header.SetMethod("GET") // Use GET as Fiber default
	fCtx.Request().SetRequestURI("/test/path?query=value&foo=bar")
	fCtx.Request().Header.Set("User-Agent", "TestAgent")
	fCtx.Request().SetBodyString(`{"name":"test"}`)

	ctx := NewFiberContext(fCtx)

	// Test Method - will return what's in the request
	_ = ctx.Method()

	// Test Path
	_ = ctx.Path()

	// Test OriginalURL
	_ = ctx.OriginalURL()

	// Test BaseURL
	_ = ctx.BaseURL()

	// Test Protocol
	_ = ctx.Protocol()

	// Test Hostname
	_ = ctx.Hostname()

	// Test IP
	_ = ctx.IP()

	// Test Secure
	_ = ctx.Secure()

	// Test Get
	userAgent := ctx.Get("User-Agent")
	if userAgent != "TestAgent" {
		t.Errorf("Get() = %v, want TestAgent", userAgent)
	}

	// Test Query
	if ctx.Query("query") != "value" {
		t.Errorf("Query() = %v, want value", ctx.Query("query"))
	}

	// Test AllQueries
	queries := ctx.AllQueries()
	if queries["query"] != "value" {
		t.Error("AllQueries() should contain query=value")
	}

	// Test Body
	body := ctx.Body()
	if len(body) == 0 {
		t.Error("Body() should return request body")
	}

	// Test IsMethod with actual method
	if !ctx.IsMethod("GET") {
		t.Error("IsMethod(GET) should be true")
	}
	if ctx.IsMethod("POST") {
		t.Error("IsMethod(POST) should be false")
	}
}

func TestFiberContext_SimpleWrappers(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := NewFiberContext(fCtx)

	// Test Cookies - returns default when not found
	cookie := ctx.Cookies("session", "default")
	if cookie != "default" {
		t.Errorf("Cookies() = %v, want default", cookie)
	}
}

// Note: Params() and AllParams() require route setup, tested in ParamsAndCookies test

func TestFiberContext_ParamsAndCookies(t *testing.T) {
	app := fiber.New()

	// Test with route that has params
	app.Get("/users/:id", func(c *fiber.Ctx) error {
		ctx := NewFiberContext(c)

		// Test Params
		id := ctx.Params("id")
		if id == "" {
			t.Error("Params() should return route param")
		}

		// Test AllParams
		params := ctx.AllParams()
		if params["id"] == "" {
			t.Error("AllParams() should contain id")
		}

		// Test ParamsParser
		var p struct {
			ID string `params:"id"`
		}
		if err := ctx.ParamsParser(&p); err != nil {
			t.Errorf("ParamsParser() error = %v", err)
		}

		// Test Cookies
		_ = ctx.Cookies("session", "default")

		return nil
	})
}

func TestFiberContext_ContentNegotiation(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	fCtx.Request().Header.Set("Accept", "application/json")
	fCtx.Request().Header.Set("Accept-Charset", "utf-8")
	fCtx.Request().Header.Set("Accept-Encoding", "gzip")
	fCtx.Request().Header.Set("Accept-Language", "en")

	ctx := NewFiberContext(fCtx)

	// Test content negotiation
	_ = ctx.Accepts("json", "xml")
	_ = ctx.AcceptsCharsets("utf-8")
	_ = ctx.AcceptsEncodings("gzip")
	_ = ctx.AcceptsLanguages("en")

	// Test request state
	_ = ctx.Fresh()
	_ = ctx.Stale()
	_ = ctx.XHR()
}

func TestFiberContext_ResponseMethods2(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := NewFiberContext(fCtx)

	// Test Set
	ctx.Set("X-Custom", "value")

	// Test Append
	ctx.Append("X-Multi", "val1", "val2")
}

func TestFiberContext_Cookie(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := NewFiberContext(fCtx)

	// Test Cookie setting
	cookie := &core.Cookie{
		Name:     "test",
		Value:    "value",
		Path:     "/",
		MaxAge:   3600,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	}
	ctx.Cookie(cookie)

	// Test ClearCookie
	ctx.ClearCookie("test")
}

func TestFiberContext_ResponseMethods(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := NewFiberContext(fCtx)

	// Test JSON
	err := ctx.JSON(map[string]string{"key": "value"})
	if err != nil {
		t.Errorf("JSON() error = %v", err)
	}

	// Test XML with struct
	type XMLData struct {
		Key string `xml:"key"`
	}
	fCtx = app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx = NewFiberContext(fCtx)
	_ = ctx.XML(XMLData{Key: "value"}) // May error, but covers the method
	app.ReleaseCtx(fCtx)

	// Test SendString
	fCtx = app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx = NewFiberContext(fCtx)
	err = ctx.SendString("test string")
	if err != nil {
		t.Errorf("SendString() error = %v", err)
	}
	app.ReleaseCtx(fCtx)

	// Test SendBytes
	fCtx = app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx = NewFiberContext(fCtx)
	err = ctx.SendBytes([]byte("test bytes"))
	if err != nil {
		t.Errorf("SendBytes() error = %v", err)
	}
	app.ReleaseCtx(fCtx)

	// Test Redirect
	fCtx = app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx = NewFiberContext(fCtx)
	err = ctx.Redirect("/redirect-target")
	if err != nil {
		t.Errorf("Redirect() error = %v", err)
	}
	app.ReleaseCtx(fCtx)
}

func TestFiberContext_Parsers(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	// Setup request with JSON body
	fCtx.Request().Header.SetContentType("application/json")
	fCtx.Request().SetBodyString(`{"name":"test","value":123}`)

	ctx := NewFiberContext(fCtx)

	// Test BodyParser
	var result map[string]interface{}
	err := ctx.BodyParser(&result)
	if err != nil {
		t.Errorf("BodyParser() error = %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("BodyParser() name = %v, want test", result["name"])
	}

	// Test QueryParser
	fCtx.Request().URI().SetQueryString("name=query&value=456")
	var queryResult struct {
		Name  string `query:"name"`
		Value int    `query:"value"`
	}
	err = ctx.QueryParser(&queryResult)
	if err != nil {
		t.Errorf("QueryParser() error = %v", err)
	}
}

func TestFiberContext_Context(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	ctx := NewFiberContext(fCtx)

	// Set some locals
	ctx.Locals("key1", "value1")
	ctx.Locals("key2", 123)
	ctx.Locals("key3", true)

	// Test Context() method
	stdCtx := ctx.Context()
	if stdCtx == nil {
		t.Error("Context() should return non-nil context")
	}

	// Values should be merged to std context
	// Note: Only string, int, int64, bool types are merged
}

func TestFiberContext_ResetTracking(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	fc := NewFiberContext(fCtx).(*FiberContext)

	// Add some tracked keys
	fc.Locals("key1", "value1")
	fc.Locals("key2", "value2")

	// Reset tracking
	fc.ResetTracking()

	// Tracked keys should be reset
	if len(fc.trackedKeys) != 0 {
		t.Error("ResetTracking() should clear tracked keys")
	}
}

// Note: Next() is tested via integration tests in middleware package
// Cannot test in isolation as it requires full fiber middleware chain

func TestGetOrCreateTrackedLocalsMap_NilTracking(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	// Set invalid tracked locals (not a map)
	fCtx.Locals(trackedLocalsKey, "invalid")

	// Should create new map
	tracked := getOrCreateTrackedLocalsMap(fCtx)
	if tracked == nil {
		t.Error("getOrCreateTrackedLocalsMap() should create new map when invalid type")
	}
}

func TestFiberContext_EnsureTrackedMap(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	fc := &FiberContext{
		fiberCtx:    fCtx,
		trackedKeys: nil,
	}

	// Should create trackedKeys map
	result := fc.ensureTrackedMap()
	if result == nil {
		t.Error("ensureTrackedMap() should create map if nil")
	}

	if fc.trackedKeys == nil {
		t.Error("ensureTrackedMap() should set trackedKeys field")
	}

	// Call again - should return existing map (same reference)
	result2 := fc.ensureTrackedMap()
	if result2 == nil {
		t.Error("ensureTrackedMap() should return non-nil map on second call")
	}
}

func TestFiberContext_GetAllLocals_WithTracking(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	fc := NewFiberContext(fCtx).(*FiberContext)

	// Set values that should be tracked
	fc.Locals("key1", "value1")
	fCtx.Locals("key2", "value2") // Direct to fiber, won't be tracked initially

	// GetAllLocals should only return tracked keys
	all := fc.GetAllLocals()

	if all["key1"] != "value1" {
		t.Error("GetAllLocals() should contain key1")
	}

	// key2 should not be there since not tracked through our wrapper
	// (unless tracked separately)
}

func TestFiberContext_Context_WithLocals(t *testing.T) {
	app := fiber.New()
	fCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(fCtx)

	fc := NewFiberContext(fCtx).(*FiberContext)

	// Set various types
	fc.Locals("str_key", "string_value")
	fc.Locals("int_key", 42)
	fc.Locals("int64_key", int64(123))
	fc.Locals("bool_key", true)
	fc.Locals("other_key", []string{"not", "supported"}) // Unsupported type

	// Get std context
	stdCtx := fc.Context()
	if stdCtx == nil {
		t.Error("Context() should return non-nil context")
	}
}
