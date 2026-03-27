package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/core/mocks"
	"github.com/anthanhphan/gosdk/tracing"
	"go.uber.org/mock/gomock"
)

func TestMiddleware_SkipForPathPrefixes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseMiddleware := func(ctx core.Context) error {
		return errors.New("middleware executed")
	}

	skipper := SkipForPathPrefixes(baseMiddleware, "/api/public", "/health")

	ctx1 := mocks.NewMockContext(ctrl)
	ctx1.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
	ctx1.EXPECT().Path().Return("/health/check").AnyTimes()
	ctx1.EXPECT().Next().Return(nil).AnyTimes()
	if err := skipper(ctx1); err != nil {
		t.Errorf("expected no error (middleware skipped), got: %v", err)
	}

	ctx2 := mocks.NewMockContext(ctrl)
	ctx2.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
	ctx2.EXPECT().Path().Return("/api/private").AnyTimes()
	if err := skipper(ctx2); err == nil || err.Error() != "middleware executed" {
		t.Errorf("expected middleware error, got: %v", err)
	}
}

func TestMiddleware_Redaction(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer secret",
		"Accept":        "application/json",
		"X-Api-Key":     "my-key-123",
	}

	sanitized := SanitizeHeaders(headers)
	if sanitized["Authorization"] != "[REDACTED]" {
		t.Error("expected Authorization to be redacted")
	}
	if sanitized["Accept"] != "application/json" {
		t.Error("expected Accept to remain intact")
	}

	// Test nil map
	if SanitizeHeaders(nil) != nil {
		t.Error("expected nil")
	}

	val := SanitizeHeaderValue("Set-Cookie", "session=123")
	if val != "[REDACTED]" {
		t.Error("expected set-cookie to be redacted")
	}

	valOk := SanitizeHeaderValue("Content-Type", "text/html")
	if valOk != "text/html" {
		t.Error("expected content-type to be intact")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := mocks.NewMockContext(ctrl)
	ctx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
	ctx.EXPECT().Get(gomock.Any(), gomock.Any()).Return("Bearer secret").AnyTimes()

	ctxHeaders := SanitizeHeadersFromContext(ctx)
	if ctxHeaders["Authorization"] != "[REDACTED]" {
		t.Error("expected Authorization to be redacted from context")
	}
}

func TestMiddleware_TracingMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := tracing.NewNoopClient()
	mw := TracingMiddleware(client)

	ctx := mocks.NewMockContext(ctrl)
	ctx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
	ctx.EXPECT().Get(gomock.Any(), gomock.Any()).Return("").AnyTimes()
	ctx.EXPECT().Method().Return(http.MethodGet).AnyTimes()
	ctx.EXPECT().OriginalURL().Return("http://localhost/api").AnyTimes()
	ctx.EXPECT().Protocol().Return("http").AnyTimes()
	ctx.EXPECT().Hostname().Return("localhost").AnyTimes()
	ctx.EXPECT().IP().Return("127.0.0.1").AnyTimes()
	ctx.EXPECT().Context().Return(context.Background()).AnyTimes()
	ctx.EXPECT().RoutePath().Return("/api").AnyTimes()

	ctx.EXPECT().SetContext(gomock.Any()).AnyTimes()
	ctx.EXPECT().Locals(gomock.Any(), gomock.Any()).AnyTimes()
	ctx.EXPECT().Set(gomock.Any(), gomock.Any()).AnyTimes()

	ctx.EXPECT().ResponseStatusCode().Return(200).AnyTimes()

	testErr := errors.New("handler error")
	ctx.EXPECT().Next().Return(testErr).AnyTimes()

	err := mw(ctx)
	if err == nil || err.Error() != "handler error" {
		t.Errorf("expected handler error, got: %v", err)
	}
}
