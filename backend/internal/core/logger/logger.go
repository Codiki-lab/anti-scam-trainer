// Package logger creates the shared Zap logger and exposes it to HTTP handlers.
package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey struct{}

type Logger struct {
	*zap.Logger
	file *os.File
}

// New creates a development-format logger that writes to standard output and a file.
func New(level, folder string) (*Logger, error) {
	var parsedLevel zap.AtomicLevel
	if err := parsedLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return nil, fmt.Errorf("create log folder: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(folder, time.Now().UTC().Format("2006-01-02T15-04-05.00000")+" logs"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.00000")
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), parsedLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(file), parsedLevel),
	)
	return &Logger{Logger: zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), file: file}, nil
}

// WithContext attaches a logger to a request context.
func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// FromContext returns the request logger, or the process default when absent.
func FromContext(ctx context.Context) *Logger {
	if log, ok := ctx.Value(contextKey{}).(*Logger); ok {
		return log
	}
	return &Logger{Logger: zap.NewNop()}
}

func (logger *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{Logger: logger.Logger.With(fields...), file: logger.file}
}

func (logger *Logger) Close() error {
	if logger.file != nil {
		if err := logger.file.Close(); err != nil {
			return fmt.Errorf("close log file: %w", err)
		}
	}
	return nil
}
