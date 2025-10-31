// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package logger

import (
	"fmt"

	"github.com/anthanhphan/gosdk/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func buildZapConfig(config *Config) zap.Config {
	// Get zap config by mode
	zapConfig := getZapConfigByMode(config)

	// Set log level
	zapConfig.Level = zap.NewAtomicLevelAt(logLevelMap[config.LogLevel])

	// Set log encoding
	zapConfig.Encoding = string(config.LogEncoding)

	// Set disable stacktrace and caller
	zapConfig.DisableStacktrace = config.DisableStacktrace
	zapConfig.DisableCaller = config.DisableCaller

	// Set encoder config
	zapConfig.EncoderConfig = buildEncoder(config)

	// Set log output paths if it's not empty
	if len(config.LogOutputPaths) != 0 {
		zapConfig.OutputPaths = config.LogOutputPaths
	}

	return zapConfig
}

func buildEncoder(config *Config) zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:    LogEncoderMessageKey,
		TimeKey:       LogEncoderTimeKey,
		LevelKey:      LogEncoderLevelKey,
		FunctionKey:   "", // Disable func_caller field
		NameKey:       LogEncoderNameKey,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeCaller:  getCallerEncoder,
		EncodeLevel:   getEncodeLevel(config),
		CallerKey:     getCallerKey(config),
		StacktraceKey: getStacktraceKey(config),
	}
}

func getZapConfigByMode(config *Config) zap.Config {
	// If development mode is enabled, use development config
	if config.IsDevelopment {
		return zap.NewDevelopmentConfig()
	}
	return zap.NewProductionConfig()
}

func getCallerKey(config *Config) string {
	if config.DisableCaller {
		return ""
	}
	return LogEncoderCallerKey
}

func getStacktraceKey(config *Config) string {
	if config.DisableStacktrace {
		return ""
	}
	return LogEncoderStacktraceKey
}

func getEncodeLevel(config *Config) zapcore.LevelEncoder {
	// If encoding is console, use color level encoder
	if config.LogEncoding == EncodingConsole {
		return zapcore.LowercaseColorLevelEncoder
	}
	return zapcore.LowercaseLevelEncoder
}

func getCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s:%d", utils.GetShortPath(caller.File), caller.Line))
}
