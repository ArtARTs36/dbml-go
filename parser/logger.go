package parser

import (
	"context"
	"fmt"
	"log/slog"
)

type Logger func(ctx context.Context, message string, params map[string]any)

var NoopLogger Logger = func(_ context.Context, _ string, _ map[string]any) {}

var SlogLogger = func(level slog.Level) Logger {
	return func(ctx context.Context, message string, params map[string]any) {
		l := slog.Default()

		for k, v := range params {
			l = l.With(slog.Any(k, v))
		}

		l.Log(ctx, level, fmt.Sprintf("[go-dbml] %s", message))
	}
}
