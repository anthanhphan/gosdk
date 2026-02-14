// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import (
	"context"
	"encoding/json"
	"errors"
)

// MockContext is a simple mock implementation of Context for testing
type MockContext struct {
	params          map[string]string
	queries         map[string]string
	headers         map[string]string
	bodyData        []byte
	bodyParseError  string
	locals          map[string]any
	statusCode      int
	responseData    any
	method          string
	path            string
	useProperStatus bool
}

func NewMockContext() *MockContext {
	return &MockContext{
		params:          make(map[string]string),
		queries:         make(map[string]string),
		headers:         make(map[string]string),
		locals:          make(map[string]any),
		statusCode:      200,
		useProperStatus: true,
	}
}

// Helper methods for setup
func (m *MockContext) AddParam(key, value string) {
	m.params[key] = value
}

func (m *MockContext) AddQuery(key, value string) {
	m.queries[key] = value
}

func (m *MockContext) SetBodyJSON(data any) {
	m.bodyData, _ = json.Marshal(data)
}

func (m *MockContext) SetBodyParseError(err string) {
	m.bodyParseError = err
}

// RequestInfo implementation
func (m *MockContext) Method() string      { return m.method }
func (m *MockContext) Path() string        { return m.path }
func (m *MockContext) RoutePath() string   { return m.path }
func (m *MockContext) OriginalURL() string { return m.path }
func (m *MockContext) BaseURL() string     { return "http://localhost" }
func (m *MockContext) Protocol() string    { return "http" }
func (m *MockContext) Hostname() string    { return "localhost" }
func (m *MockContext) IP() string          { return "127.0.0.1" }
func (m *MockContext) Secure() bool        { return false }

// HeaderManager implementation
func (m *MockContext) Get(key string, defaultValue ...string) string {
	if val, ok := m.headers[key]; ok {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func (m *MockContext) Set(key, value string) {
	m.headers[key] = value
}

func (m *MockContext) Append(field string, values ...string) {
	// Simple implementation
	for _, v := range values {
		m.headers[field] = v
	}
}

// ParamGetter implementation
func (m *MockContext) Params(key string, defaultValue ...string) string {
	if val, ok := m.params[key]; ok {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func (m *MockContext) AllParams() map[string]string {
	return m.params
}

func (m *MockContext) ParamsParser(out any) error {
	data, _ := json.Marshal(m.params)
	return json.Unmarshal(data, out)
}

// QueryGetter implementation
func (m *MockContext) Query(key string, defaultValue ...string) string {
	if val, ok := m.queries[key]; ok {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func (m *MockContext) AllQueries() map[string]string {
	return m.queries
}

func (m *MockContext) QueryParser(out any) error {
	data, _ := json.Marshal(m.queries)
	return json.Unmarshal(data, out)
}

// BodyParser implementation
func (m *MockContext) Body() []byte {
	return m.bodyData
}

func (m *MockContext) BodyParser(out any) error {
	if m.bodyParseError != "" {
		return errors.New(m.bodyParseError)
	}
	return json.Unmarshal(m.bodyData, out)
}

// CookieManager implementation
func (m *MockContext) Cookies(key string, defaultValue ...string) string {
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func (m *MockContext) Cookie(cookie *Cookie) {}

func (m *MockContext) ClearCookie(key ...string) {}

// ResponseWriter implementation
func (m *MockContext) Status(status int) Context {
	m.statusCode = status
	return m
}

func (m *MockContext) ResponseStatusCode() int {
	return m.statusCode
}

func (m *MockContext) JSON(data any) error {
	m.responseData = data
	return nil
}

func (m *MockContext) XML(data any) error {
	m.responseData = data
	return nil
}

func (m *MockContext) SendString(s string) error {
	m.responseData = s
	return nil
}

func (m *MockContext) SendBytes(b []byte) error {
	m.responseData = b
	return nil
}

func (m *MockContext) Redirect(location string, status ...int) error {
	if len(status) > 0 {
		m.statusCode = status[0]
	}
	return nil
}

// ContentNegotiator implementation
func (m *MockContext) Accepts(offers ...string) string {
	if len(offers) > 0 {
		return offers[0]
	}
	return ""
}

func (m *MockContext) AcceptsCharsets(offers ...string) string {
	if len(offers) > 0 {
		return offers[0]
	}
	return ""
}

func (m *MockContext) AcceptsEncodings(offers ...string) string {
	if len(offers) > 0 {
		return offers[0]
	}
	return ""
}

func (m *MockContext) AcceptsLanguages(offers ...string) string {
	if len(offers) > 0 {
		return offers[0]
	}
	return ""
}

// RequestState implementation
func (m *MockContext) Fresh() bool { return false }
func (m *MockContext) Stale() bool { return true }
func (m *MockContext) XHR() bool   { return false }

// LocalsStorage implementation
func (m *MockContext) Locals(key string, value ...any) any {
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}

func (m *MockContext) GetAllLocals() map[string]any {
	return m.locals
}

// ShorthandResponder implementation
func (m *MockContext) OK(data any) error {
	m.statusCode = StatusOK
	m.responseData = data
	return nil
}

func (m *MockContext) Created(data any) error {
	m.statusCode = StatusCreated
	m.responseData = data
	return nil
}

func (m *MockContext) NoContent() error {
	m.statusCode = StatusNoContent
	return nil
}

func (m *MockContext) BadRequestMsg(message string) error {
	m.statusCode = StatusBadRequest
	m.responseData = message
	return nil
}

func (m *MockContext) UnauthorizedMsg(message string) error {
	m.statusCode = StatusUnauthorized
	m.responseData = message
	return nil
}

func (m *MockContext) ForbiddenMsg(message string) error {
	m.statusCode = StatusForbidden
	m.responseData = message
	return nil
}

func (m *MockContext) NotFoundMsg(message string) error {
	m.statusCode = StatusNotFound
	m.responseData = message
	return nil
}

func (m *MockContext) InternalErrorMsg(message string) error {
	m.statusCode = StatusInternalServerError
	m.responseData = message
	return nil
}

// Other Context methods
func (m *MockContext) Next() error {
	return nil
}

func (m *MockContext) Context() context.Context {
	return context.Background()
}

func (m *MockContext) IsMethod(method string) bool {
	return m.method == method
}

func (m *MockContext) RequestID() string {
	return "mock-request-id"
}

func (m *MockContext) UseProperHTTPStatus() bool {
	return m.useProperStatus
}
