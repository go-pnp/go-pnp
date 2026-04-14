package pnphttpserversentry

import (
	"context"

	"github.com/getsentry/sentry-go"
)

// startSpanOrTransaction creates a child span if a transaction already exists
// in the context, otherwise starts a new root transaction.
func startSpanOrTransaction(ctx context.Context, name, op, sentryTrace, baggage string) *sentry.Span {
	if span := sentry.SpanFromContext(ctx); span != nil {
		return sentry.StartSpan(ctx, op, sentry.WithDescription(name))
	}

	return sentry.StartTransaction(ctx, name,
		sentry.WithOpName(op),
		sentry.ContinueFromHeaders(sentryTrace, baggage),
	)
}
