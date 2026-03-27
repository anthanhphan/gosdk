package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/core/mocks"
	enginemocks "github.com/anthanhphan/gosdk/orianna/http/engine/mocks"
	"github.com/anthanhphan/gosdk/orianna/http/routing"
	"github.com/anthanhphan/gosdk/orianna/shared/health"
	"github.com/anthanhphan/gosdk/tracing"
	"go.uber.org/mock/gomock"
)

func TestServer_HooksMiddleware(t *testing.T) {
	hooks := core.NewHooks()

	reqCalled := false
	resCalled := false
	errCalled := false

	hooks.AddOnRequest(func(ctx core.Context) {
		reqCalled = true
	})
	hooks.AddOnResponse(func(ctx core.Context, code int, latency time.Duration) {
		resCalled = true
	})
	hooks.AddOnError(func(ctx core.Context, err error) {
		errCalled = true
	})

	mw := hooksMiddleware(hooks)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := mocks.NewMockContext(ctrl)
	ctx.EXPECT().Accepts(gomock.Any(), gomock.Any()).Return("application/json").AnyTimes()
	ctx.EXPECT().Next().Return(errors.New("handler error"))
	ctx.EXPECT().ResponseStatusCode().Return(500)

	err := mw(ctx)
	if err == nil || err.Error() != "handler error" {
		t.Errorf("expected handler error")
	}

	if !reqCalled || !resCalled || !errCalled {
		t.Error("expected hooks to be called")
	}
}

type dummyChecker struct{}

func (d dummyChecker) Check(ctx context.Context) health.HealthCheck { return health.HealthCheck{} }
func (d dummyChecker) Name() string                                 { return "dummy" }

func TestServer_Options(t *testing.T) {
	hc := dummyChecker{}
	s, err := NewServer(&configuration.Config{Port: 0, ServiceName: "test"},
		WithTracing(tracing.NewNoopClient()),
		WithHealthChecker(hc),
	)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if s.healthManager == nil {
		t.Error("expected health checker to be set and wrapped in manager")
	}
	if s.tracingClient == nil {
		t.Error("expected tracing client to be set")
	}
}

func TestServer_StartShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	engineMock := enginemocks.NewMockServerEngine(ctrl)
	engineMock.EXPECT().RegisterRoutes(gomock.Any()).Return(nil).AnyTimes()
	engineMock.EXPECT().RegisterGroup(gomock.Any()).Return(nil).AnyTimes()
	engineMock.EXPECT().Start().Return(nil).AnyTimes()
	engineMock.EXPECT().Shutdown(gomock.Any()).Return(nil).AnyTimes()
	engineMock.EXPECT().SetupGlobalMiddlewares(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	engineMock.EXPECT().SetupLoggingMiddleware(gomock.Any(), gomock.Any()).AnyTimes()
	engineMock.EXPECT().Use(gomock.Any()).AnyTimes()

	s, err := NewServer(&configuration.Config{Port: 0, ServiceName: "test"}, WithServerEngine(engineMock))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	route := routing.NewRoute("/test").Method(core.GET).Handler(func(c core.Context) error { return nil }).Build()
	err = s.RegisterRoutes(*route)
	if err != nil {
		t.Errorf("failed to register route: %v", err)
	}

	groupRoute := routing.NewRoute("/testg").Method(core.GET).Handler(func(c core.Context) error { return nil }).Build()
	group := routing.NewGroupRoute("/api").Route(groupRoute).Build()
	err = s.RegisterGroup(*group)
	if err != nil {
		t.Errorf("failed to register group: %v", err)
	}

	go func() {
		_ = s.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	err = s.Shutdown(context.Background())
	if err != nil {
		t.Errorf("failed to shutdown gracefully: %v", err)
	}
}

func TestServer_RunShortcut(t *testing.T) {
	defer func() {
		_ = recover()
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	engineMock := enginemocks.NewMockServerEngine(ctrl)
	engineMock.EXPECT().Start().Return(nil).AnyTimes()
	engineMock.EXPECT().Shutdown(gomock.Any()).Return(nil).AnyTimes()
	engineMock.EXPECT().SetupGlobalMiddlewares(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	engineMock.EXPECT().SetupLoggingMiddleware(gomock.Any(), gomock.Any()).AnyTimes()
	engineMock.EXPECT().Use(gomock.Any()).AnyTimes()

	s, _ := NewServer(&configuration.Config{Port: 0, ServiceName: "test"}, WithServerEngine(engineMock))
	if s == nil {
		t.Skip("server init failed")
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = s.Shutdown(context.Background())
	}()

	_ = s.Run()
}
