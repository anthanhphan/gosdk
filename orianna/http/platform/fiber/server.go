// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package fiber

import (
	"context"
	"fmt"
	"net"

	"github.com/anthanhphan/gosdk/jcodec"
	"github.com/anthanhphan/gosdk/orianna/http/configuration"
	"github.com/anthanhphan/gosdk/orianna/http/core"
	"github.com/anthanhphan/gosdk/orianna/http/engine"
	"github.com/gofiber/fiber/v3"
)

// Compile-time interface compliance check
var _ engine.ServerEngine = (*ServerAdapter)(nil)

// ctxAdapterKey is the fiber Locals key for reusing ContextAdapter within a request.
// This must be unexported (lowercase) and unique to avoid collision with user keys.
const ctxAdapterKey = "__orianna_ctx"

// ServerAdapter implements ServerEngine using Fiber v3
type ServerAdapter struct {
	app    *fiber.App
	config *configuration.Config
	router engine.RouterEngine
}

// NewServerAdapter creates a new Fiber server adapter
func NewServerAdapter(conf *configuration.Config) (*ServerAdapter, error) {
	readTimeout := configuration.DefaultReadTimeout
	if conf.ReadTimeout != nil {
		readTimeout = *conf.ReadTimeout
	}

	writeTimeout := configuration.DefaultWriteTimeout
	if conf.WriteTimeout != nil {
		writeTimeout = *conf.WriteTimeout
	}

	idleTimeout := configuration.DefaultIdleTimeout
	if conf.IdleTimeout != nil {
		idleTimeout = *conf.IdleTimeout
	}

	bodyLimit := conf.MaxBodySize
	if bodyLimit <= 0 {
		bodyLimit = configuration.DefaultMaxBodySize
	}

	concurrency := conf.MaxConcurrentConnections
	if concurrency <= 0 {
		concurrency = configuration.DefaultMaxConcurrentConnections
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		BodyLimit:    bodyLimit,
		Concurrency:  concurrency,
		JSONEncoder:  jcodec.Marshal,
		JSONDecoder:  jcodec.Unmarshal,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			errResp := core.NewErrorResponse("INTERNAL_ERROR", core.StatusInternalServerError, err.Error())
			return c.Status(core.StatusInternalServerError).JSON(errResp)
		},
	})

	adapter := &ServerAdapter{
		app:    app,
		config: conf,
	}

	adapter.router = newRouterAdapterWithConfig(app, conf)
	return adapter, nil
}

// Start starts the HTTP server on the configured port
func (s *ServerAdapter) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	return s.app.Listener(ln)
}

// Shutdown gracefully shuts down the server
func (s *ServerAdapter) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// Use adds domain-level middleware to the server
func (s *ServerAdapter) Use(middleware ...core.Middleware) {
	for _, mw := range middleware {
		s.app.Use(convertToFiberMiddlewareWithConfig(mw, s.config))
	}
}
