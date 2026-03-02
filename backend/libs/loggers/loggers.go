package loggers

import (
	"log/slog"
	"os"
)

type HandlerType string

const (
	JSON HandlerType = "json"
	Text HandlerType = "text"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Config struct {
	Level   LogLevel
	Handler HandlerType
}

type LoggerInterface interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func New(cfg Config) *slog.Logger {
	return newWithConfig(cfg)
}

func newWithConfig(cfg Config) *slog.Logger {
	var handler slog.Handler

	level := slog.LevelInfo
	switch cfg.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Default handler to Text if empty
	if cfg.Handler == "" {
		cfg.Handler = Text
	}

	switch cfg.Handler {
	case JSON:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case Text:
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
