// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package middleware

import "github.com/anthanhphan/gosdk/logger"

// defaultLog is the package-level logger for all middleware that don't receive
// an explicit logger instance. Used by Recover, SlowRequestDetector, etc.
var defaultLog = logger.NewLoggerWithFields(logger.String("package", "middleware"))
