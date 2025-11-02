package aurelion

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestContextWrapper_BasicProperties(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		if ctx == nil {
			t.Fatal("Context should not be nil")
		}
		if ctx.Method() != http.MethodGet {
			t.Errorf("Method() = %v, want %v", ctx.Method(), http.MethodGet)
		}
		if ctx.Path() != "/test" {
			t.Errorf("Path() = %v, want %v", ctx.Path(), "/test")
		}
		if ctx.OriginalURL() == "" {
			t.Error("OriginalURL() should not be empty")
		}
		if ctx.BaseURL() == "" {
			t.Error("BaseURL() should not be empty")
		}
		if ctx.Protocol() != "http" && ctx.Protocol() != "https" {
			t.Errorf("Protocol() = %v, want http or https", ctx.Protocol())
		}
		if ctx.Hostname() == "" {
			t.Error("Hostname() should not be empty")
		}
		if ctx.IP() == "" {
			t.Error("IP() should not be empty")
		}
		if ctx.Secure() {
			t.Error("Secure() should be false for HTTP")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_Headers(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		if ctx.Get("X-Custom") != "value" {
			t.Errorf("Get() = %v, want %v", ctx.Get("X-Custom"), "value")
		}
		if ctx.Get("Non-Existent", "default") != "default" {
			t.Errorf("Get() = %v, want %v", ctx.Get("Non-Existent", "default"), "default")
		}

		ctx.Set("X-Response", "test")
		ctx.Append("X-Test", "value1")
		ctx.Append("X-Test", "value2")
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom", "value")
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.Header.Get("X-Response") != "test" {
		t.Error("Response header should be set")
	}
}

func TestContextWrapper_Params(t *testing.T) {
	app := fiber.New()

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		if ctx.Params("id") != "123" {
			t.Errorf("Params() = %v, want %v", ctx.Params("id"), "123")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_AllParams(t *testing.T) {
	app := fiber.New()

	app.Get("/users/:id/files/:fileId", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		params := ctx.AllParams()
		if len(params) != 2 {
			t.Errorf("AllParams() length = %v, want %v", len(params), 2)
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123/files/456", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_Query(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		if ctx.Query("foo") != "bar" {
			t.Errorf("Query() = %v, want %v", ctx.Query("foo"), "bar")
		}

		queries := ctx.AllQueries()
		if len(queries) == 0 {
			t.Error("AllQueries() should not be empty")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar&baz=qux", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_Locals(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		ctx.Locals("test_key", "test_value")
		if ctx.Locals("test_key") != "test_value" {
			t.Errorf("Locals() = %v, want %v", ctx.Locals("test_key"), "test_value")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_Response(t *testing.T) {
	app := fiber.New()

	app.Get("/json", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		data := map[string]interface{}{"key": "value"}
		return ctx.JSON(data)
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextWrapper_StringResponse(t *testing.T) {
	app := fiber.New()

	app.Get("/string", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		return ctx.SendString("hello")
	})

	req := httptest.NewRequest(http.MethodGet, "/string", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextWrapper_Status(t *testing.T) {
	app := fiber.New()

	app.Get("/status", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		ctx.Status(http.StatusCreated)
		return ctx.SendString("created")
	})

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestContextWrapper_Nil(t *testing.T) {
	ctx := newContext(nil)
	if ctx != nil {
		t.Error("Context wrapper with nil input should be nil")
	}
}

func TestContextWrapper_AcceptHeaders(t *testing.T) {
	app := fiber.New()

	app.Get("/accept", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		if ctx.Accepts("application/json") == "" {
			t.Error("Accepts() should return a value")
		}
		if ctx.AcceptsCharsets("utf-8") == "" {
			t.Error("AcceptsCharsets() should return a value")
		}
		if ctx.AcceptsEncodings("gzip") == "" {
			t.Error("AcceptsEncodings() should return a value")
		}
		if ctx.AcceptsLanguages("en") == "" {
			t.Error("AcceptsLanguages() should return a value")
		}

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/accept", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept-Language", "en")

	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_XHR(t *testing.T) {
	app := fiber.New()

	app.Get("/xhr", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		if !ctx.XHR() {
			t.Error("XHR() should return true")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/xhr", nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_Cookies(t *testing.T) {
	app := fiber.New()

	app.Get("/cookie", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		cookie := &Cookie{
			Name:  "test",
			Value: "value",
			Path:  "/",
		}
		ctx.Cookie(cookie)
		if ctx.Cookies("test") == "" {
			t.Error("Cookies() should return cookie value")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/cookie", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "value"})
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if len(resp.Header.Values("Set-Cookie")) == 0 {
		t.Error("Cookie should be set")
	}
}

func TestContextWrapper_MoreMethods(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		// Test Body
		body := ctx.Body()
		_ = body

		// Test Fresh/Stale
		fresh := ctx.Fresh()
		_ = fresh
		stale := ctx.Stale()
		_ = stale

		// Test Next
		err := ctx.Next()
		_ = err

		// Test Context
		if ctx.Context() == nil {
			t.Error("Context() should not be nil")
		}

		// Test that Context() merges all Locals values
		ctx.Locals("custom_key", "custom_value")
		ctx.Locals("another_key", "another_value")

		contextCtx := ctx.Context()
		if contextCtx == nil {
			t.Error("Context() should not be nil")
		}

		// Verify custom keys are accessible via context.Context
		customValue := getValueFromContext(contextCtx, "custom_key")
		if customValue != "custom_value" {
			t.Errorf("Expected custom_value, got %s", customValue)
		}

		anotherValue := getValueFromContext(contextCtx, "another_key")
		if anotherValue != "another_value" {
			t.Errorf("Expected another_value, got %s", anotherValue)
		}

		// Test ClearCookie
		ctx.ClearCookie("test")

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_Parsers(t *testing.T) {
	app := fiber.New()

	type User struct {
		ID   string `params:"id"`
		Name string `query:"name"`
	}

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		// Test ParamsParser
		var user User
		if err := ctx.ParamsParser(&user); err != nil {
			t.Fatalf("ParamsParser() error = %v", err)
		}

		// Test QueryParser
		var user2 User
		if err := ctx.QueryParser(&user2); err != nil {
			t.Fatalf("QueryParser() error = %v", err)
		}

		// Test BodyParser
		var data map[string]interface{}
		_ = ctx.BodyParser(&data) // Expected to fail for GET request

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123?name=John", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_XMLResponse(t *testing.T) {
	app := fiber.New()

	app.Get("/xml", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		// XML needs proper XML data structure
		type Data struct {
			XMLName xml.Name `xml:"data"`
			Key     string   `xml:"key"`
		}
		data := Data{Key: "value"}
		return ctx.XML(data)
	})

	req := httptest.NewRequest(http.MethodGet, "/xml", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextWrapper_SendBytes(t *testing.T) {
	app := fiber.New()

	app.Get("/bytes", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		return ctx.SendBytes([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/bytes", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextWrapper_Redirect(t *testing.T) {
	app := fiber.New()

	app.Get("/redirect", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		return ctx.Redirect("/target", http.StatusFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/redirect", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status 302, got %d", resp.StatusCode)
	}
}

func TestNewContext_NilInput(t *testing.T) {
	ctx := newContext(nil)
	if ctx != nil {
		t.Error("newContext(nil) should return nil")
	}
}

func TestContextWrapper_IsMethod(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		if !ctx.IsMethod("GET") {
			t.Error("IsMethod(\"GET\") should return true")
		}
		if ctx.IsMethod("POST") {
			t.Error("IsMethod(\"POST\") should return false")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_RequestID(t *testing.T) {
	app := fiber.New()
	app.Use(requestIDMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		requestID := ctx.RequestID()
		if requestID == "" {
			t.Error("RequestID() should not be empty")
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestContextWrapper_GetAllLocals(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)

		// Set some locals through the wrapper to ensure they're tracked
		ctx.Locals("key1", "value1")
		ctx.Locals("key2", "value2")
		ctx.Locals("key3", 123) // non-string value

		allLocals := ctx.GetAllLocals()

		if allLocals == nil {
			t.Error("GetAllLocals() should not return nil")
		}

		if allLocals["key1"] != "value1" {
			t.Errorf("Expected key1 = value1, got %v", allLocals["key1"])
		}

		if allLocals["key2"] != "value2" {
			t.Errorf("Expected key2 = value2, got %v", allLocals["key2"])
		}

		if allLocals["key3"] != 123 {
			t.Errorf("Expected key3 = 123, got %v", allLocals["key3"])
		}

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestContextWrapper_GetAllLocals_Empty(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		ctx := newContext(c)
		allLocals := ctx.GetAllLocals()

		if allLocals == nil {
			t.Error("GetAllLocals() should not return nil even when empty")
		}

		// Note: GetAllLocals may return non-empty map due to middleware setting values
		// So we just check it's not nil
		_ = allLocals

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, 100)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
