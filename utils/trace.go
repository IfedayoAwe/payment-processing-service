package utils

import (
	"context"

	"github.com/google/uuid"
)

type traceIDKey struct{}

const TraceIDHeader = "X-Trace-ID"

func GenerateTraceID() string {
	return uuid.New().String()
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}
