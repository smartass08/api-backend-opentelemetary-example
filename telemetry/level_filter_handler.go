package telemetry

import (
	"context"
	"log/slog"
)

// LevelFilterHandler wraps another handler and filters by log level
type LevelFilterHandler struct {
	handler  slog.Handler
	minLevel slog.Level
}

func NewLevelFilterHandler(handler slog.Handler, minLevel slog.Level) *LevelFilterHandler {
	return &LevelFilterHandler{
		handler:  handler,
		minLevel: minLevel,
	}
}

func (l *LevelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= l.minLevel && l.handler.Enabled(ctx, level)
}

func (l *LevelFilterHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= l.minLevel {
		return l.handler.Handle(ctx, record)
	}
	return nil
}

func (l *LevelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelFilterHandler{
		handler:  l.handler.WithAttrs(attrs),
		minLevel: l.minLevel,
	}
}

func (l *LevelFilterHandler) WithGroup(name string) slog.Handler {
	return &LevelFilterHandler{
		handler:  l.handler.WithGroup(name),
		minLevel: l.minLevel,
	}
}