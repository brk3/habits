package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func Init(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

func InitJSON(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

func Get() *slog.Logger {
	if defaultLogger == nil {
		Init(slog.LevelInfo)
	}
	return defaultLogger
}

func With(args ...any) *slog.Logger {
	return Get().With(args...)
}

func WithGroup(name string) *slog.Logger {
	return Get().WithGroup(name)
}

func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	Get().DebugContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	Get().InfoContext(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	Get().WarnContext(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	Get().ErrorContext(ctx, msg, args...)
}