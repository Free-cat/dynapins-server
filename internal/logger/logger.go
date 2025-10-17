package logger

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func Init() {
	InitWithLevel("info")
}

func InitWithLevel(level string) {
	logLevel := parseLogLevel(level)
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getLogger() *slog.Logger {
	if Logger == nil {
		// Auto-initialize if not already initialized
		Init()
	}
	return Logger
}

func Info(msg string, args ...any) {
	getLogger().Info(msg, args...)
}

func Error(msg string, args ...any) {
	getLogger().Error(msg, args...)
}

func Warn(msg string, args ...any) {
	getLogger().Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	getLogger().Debug(msg, args...)
}
