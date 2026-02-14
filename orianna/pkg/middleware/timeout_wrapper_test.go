// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import (
	"context"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/pkg/core"
	"github.com/anthanhphan/gosdk/orianna/pkg/core/mocks"
	"go.uber.org/mock/gomock"
)

// TestTimeoutContextWrapper tests delegation and Context() override.
func TestTimeoutContextWrapper_ContextOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCtx := mocks.NewMockContext(ctrl)
	timeoutCtx := context.Background()

	wrapper := &timeoutContextWrapper{inner: mockCtx, timeoutCtx: timeoutCtx}

	// Context() should return timeoutCtx, not inner's context
	if wrapper.Context() != timeoutCtx {
		t.Error("Context() should return timeoutCtx")
	}
}

// TestTimeoutContextWrapper_StatusChaining verifies Status() returns wrapper
func TestTimeoutContextWrapper_StatusChaining(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCtx := mocks.NewMockContext(ctrl)
	timeoutCtx := context.Background()

	wrapper := &timeoutContextWrapper{inner: mockCtx, timeoutCtx: timeoutCtx}

	mockCtx.EXPECT().Status(200)
	result := wrapper.Status(200)
	if result != wrapper {
		t.Error("Status() should return wrapper for chaining")
	}
}

// TestTimeoutContextWrapper_AllDelegations tests all delegation methods of timeoutContextWrapper.
func TestTimeoutContextWrapper_AllDelegations(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCtx := mocks.NewMockContext(ctrl)
	timeoutCtx := context.Background()

	wrapper := &timeoutContextWrapper{inner: mockCtx, timeoutCtx: timeoutCtx}

	// Test all delegation methods
	mockCtx.EXPECT().Next().Return(nil)
	if err := wrapper.Next(); err != nil {
		t.Errorf("Next() error = %v", err)
	}

	mockCtx.EXPECT().Method().Return("GET")
	if wrapper.Method() != "GET" {
		t.Errorf("Method() = %s, want GET", wrapper.Method())
	}

	mockCtx.EXPECT().Path().Return("/test")
	if wrapper.Path() != "/test" {
		t.Errorf("Path() = %s, want /test", wrapper.Path())
	}

	mockCtx.EXPECT().RoutePath().Return("/test/:id")
	if wrapper.RoutePath() != "/test/:id" {
		t.Errorf("RoutePath() = %s, want /test/:id", wrapper.RoutePath())
	}

	mockCtx.EXPECT().OriginalURL().Return("/test/123?a=b")
	if wrapper.OriginalURL() != "/test/123?a=b" {
		t.Error("OriginalURL() mismatch")
	}

	mockCtx.EXPECT().BaseURL().Return("http://localhost:8080")
	if wrapper.BaseURL() != "http://localhost:8080" {
		t.Error("BaseURL() mismatch")
	}

	mockCtx.EXPECT().Protocol().Return("https")
	if wrapper.Protocol() != "https" {
		t.Error("Protocol() mismatch")
	}

	mockCtx.EXPECT().Hostname().Return("localhost")
	if wrapper.Hostname() != "localhost" {
		t.Error("Hostname() mismatch")
	}

	mockCtx.EXPECT().IP().Return("127.0.0.1")
	if wrapper.IP() != "127.0.0.1" {
		t.Error("IP() mismatch")
	}

	mockCtx.EXPECT().Secure().Return(true)
	if !wrapper.Secure() {
		t.Error("Secure() should be true")
	}

	mockCtx.EXPECT().Get("Content-Type").Return("application/json")
	if wrapper.Get("Content-Type") != "application/json" {
		t.Error("Get() mismatch")
	}

	mockCtx.EXPECT().Set("X-Custom", "value")
	wrapper.Set("X-Custom", "value")

	mockCtx.EXPECT().Append("X-Custom", "v1", "v2")
	wrapper.Append("X-Custom", "v1", "v2")

	mockCtx.EXPECT().Params("id").Return("123")
	if wrapper.Params("id") != "123" {
		t.Error("Params() mismatch")
	}

	mockCtx.EXPECT().AllParams().Return(map[string]string{"id": "123"})
	params := wrapper.AllParams()
	if params["id"] != "123" {
		t.Error("AllParams() mismatch")
	}

	mockCtx.EXPECT().ParamsParser(gomock.Any()).Return(nil)
	if err := wrapper.ParamsParser(nil); err != nil {
		t.Error("ParamsParser() error")
	}

	mockCtx.EXPECT().Query("page").Return("1")
	if wrapper.Query("page") != "1" {
		t.Error("Query() mismatch")
	}

	mockCtx.EXPECT().AllQueries().Return(map[string]string{"page": "1"})
	queries := wrapper.AllQueries()
	if queries["page"] != "1" {
		t.Error("AllQueries() mismatch")
	}

	mockCtx.EXPECT().QueryParser(gomock.Any()).Return(nil)
	if err := wrapper.QueryParser(nil); err != nil {
		t.Error("QueryParser() error")
	}

	mockCtx.EXPECT().Body().Return([]byte(`{"key":"value"}`))
	if string(wrapper.Body()) != `{"key":"value"}` {
		t.Error("Body() mismatch")
	}

	mockCtx.EXPECT().BodyParser(gomock.Any()).Return(nil)
	if err := wrapper.BodyParser(nil); err != nil {
		t.Error("BodyParser() error")
	}

	mockCtx.EXPECT().Cookies("session").Return("abc")
	if wrapper.Cookies("session") != "abc" {
		t.Error("Cookies() mismatch")
	}

	cookie := &core.Cookie{Name: "test", Value: "val"}
	mockCtx.EXPECT().Cookie(cookie)
	wrapper.Cookie(cookie)

	mockCtx.EXPECT().ClearCookie("session")
	wrapper.ClearCookie("session")

	mockCtx.EXPECT().ResponseStatusCode().Return(200)
	if wrapper.ResponseStatusCode() != 200 {
		t.Error("ResponseStatusCode() mismatch")
	}

	mockCtx.EXPECT().JSON(gomock.Any()).Return(nil)
	if err := wrapper.JSON(nil); err != nil {
		t.Error("JSON() error")
	}

	mockCtx.EXPECT().XML(gomock.Any()).Return(nil)
	if err := wrapper.XML(nil); err != nil {
		t.Error("XML() error")
	}

	mockCtx.EXPECT().SendString("hello").Return(nil)
	if err := wrapper.SendString("hello"); err != nil {
		t.Error("SendString() error")
	}

	mockCtx.EXPECT().SendBytes([]byte("hello")).Return(nil)
	if err := wrapper.SendBytes([]byte("hello")); err != nil {
		t.Error("SendBytes() error")
	}

	mockCtx.EXPECT().Redirect("/other", 302).Return(nil)
	if err := wrapper.Redirect("/other", 302); err != nil {
		t.Error("Redirect() error")
	}

	mockCtx.EXPECT().Accepts("json", "xml").Return("json")
	if wrapper.Accepts("json", "xml") != "json" {
		t.Error("Accepts() mismatch")
	}

	mockCtx.EXPECT().AcceptsCharsets("utf-8").Return("utf-8")
	if wrapper.AcceptsCharsets("utf-8") != "utf-8" {
		t.Error("AcceptsCharsets() mismatch")
	}

	mockCtx.EXPECT().AcceptsEncodings("gzip").Return("gzip")
	if wrapper.AcceptsEncodings("gzip") != "gzip" {
		t.Error("AcceptsEncodings() mismatch")
	}

	mockCtx.EXPECT().AcceptsLanguages("en").Return("en")
	if wrapper.AcceptsLanguages("en") != "en" {
		t.Error("AcceptsLanguages() mismatch")
	}

	mockCtx.EXPECT().Fresh().Return(true)
	if !wrapper.Fresh() {
		t.Error("Fresh() should be true")
	}

	mockCtx.EXPECT().Stale().Return(false)
	if wrapper.Stale() {
		t.Error("Stale() should be false")
	}

	mockCtx.EXPECT().XHR().Return(true)
	if !wrapper.XHR() {
		t.Error("XHR() should be true")
	}

	mockCtx.EXPECT().Locals("user", "john").Return("john")
	if wrapper.Locals("user", "john") != "john" {
		t.Error("Locals() mismatch")
	}

	mockCtx.EXPECT().GetAllLocals().Return(map[string]any{"user": "john"})
	locals := wrapper.GetAllLocals()
	if locals["user"] != "john" {
		t.Error("GetAllLocals() mismatch")
	}

	mockCtx.EXPECT().OK(gomock.Any()).Return(nil)
	if err := wrapper.OK(nil); err != nil {
		t.Error("OK() error")
	}

	mockCtx.EXPECT().Created(gomock.Any()).Return(nil)
	if err := wrapper.Created(nil); err != nil {
		t.Error("Created() error")
	}

	mockCtx.EXPECT().NoContent().Return(nil)
	if err := wrapper.NoContent(); err != nil {
		t.Error("NoContent() error")
	}

	mockCtx.EXPECT().BadRequestMsg("bad").Return(nil)
	if err := wrapper.BadRequestMsg("bad"); err != nil {
		t.Error("BadRequestMsg() error")
	}

	mockCtx.EXPECT().UnauthorizedMsg("unauth").Return(nil)
	if err := wrapper.UnauthorizedMsg("unauth"); err != nil {
		t.Error("UnauthorizedMsg() error")
	}

	mockCtx.EXPECT().ForbiddenMsg("forbidden").Return(nil)
	if err := wrapper.ForbiddenMsg("forbidden"); err != nil {
		t.Error("ForbiddenMsg() error")
	}

	mockCtx.EXPECT().NotFoundMsg("not found").Return(nil)
	if err := wrapper.NotFoundMsg("not found"); err != nil {
		t.Error("NotFoundMsg() error")
	}

	mockCtx.EXPECT().InternalErrorMsg("error").Return(nil)
	if err := wrapper.InternalErrorMsg("error"); err != nil {
		t.Error("InternalErrorMsg() error")
	}

	mockCtx.EXPECT().IsMethod("GET").Return(true)
	if !wrapper.IsMethod("GET") {
		t.Error("IsMethod() should be true")
	}

	mockCtx.EXPECT().RequestID().Return("req-123")
	if wrapper.RequestID() != "req-123" {
		t.Error("RequestID() mismatch")
	}

	mockCtx.EXPECT().UseProperHTTPStatus().Return(true)
	if !wrapper.UseProperHTTPStatus() {
		t.Error("UseProperHTTPStatus() should be true")
	}
}
