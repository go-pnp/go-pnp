package pnpsentryconnectrpchandling

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/getsentry/sentry-go"
	"github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := optionutil.ApplyOptions(&options{}, opts...)

	builder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	builder.Supply(options)
	builder.Provide(NewInterceptor)
	builder.Provide(pnpconnectrpchandling.InterceptorProvider(NewInterceptor))

	return builder.Build()
}

type Interceptor struct {
	client *sentry.Client
}

func NewInterceptor(client *sentry.Client, options *options) ordering.OrderedItem[connect.Interceptor] {
	return ordering.OrderedItem[connect.Interceptor]{
		Value: &Interceptor{client: client},
		Order: options.order,
	}
}

func (i Interceptor) WrapUnary(unaryFunc connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.NewHub(i.client, sentry.NewScope())
			ctx = sentry.SetHubOnContext(ctx, hub)
		} else {
			hub.PushScope()
			defer hub.PopScope()
		}
		setRequestInfo(hub.Scope(), request.Spec(), request.Header())

		span := startSpanOrTransaction(ctx, request.Spec().Procedure, "connectrpc.unary",
			request.Header().Get("sentry-trace"), request.Header().Get("baggage"),
		)
		defer span.Finish()
		ctx = span.Context()

		return unaryFunc(ctx, request)
	}
}

func (i Interceptor) WrapStreamingClient(clientFunc connect.StreamingClientFunc) connect.StreamingClientFunc {
	return clientFunc
}

func (i Interceptor) WrapStreamingHandler(handlerFunc connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.NewHub(i.client, sentry.NewScope())
			ctx = sentry.SetHubOnContext(ctx, hub)
		} else {
			hub.PushScope()
			defer hub.PopScope()
		}
		setRequestInfo(hub.Scope(), conn.Spec(), conn.RequestHeader())

		span := startSpanOrTransaction(ctx, conn.Spec().Procedure, "connectrpc.stream",
			conn.RequestHeader().Get("sentry-trace"), conn.RequestHeader().Get("baggage"),
		)
		defer span.Finish()
		ctx = span.Context()

		return handlerFunc(ctx, conn)
	}
}

func setRequestInfo(scope *sentry.Scope, spec connect.Spec, headers map[string][]string) {
	service, method := splitProcedure(spec.Procedure)
	scope.SetTag("connectrpc.procedure", spec.Procedure)
	scope.SetTag("connectrpc.service", service)
	scope.SetTag("connectrpc.method", method)
	scope.SetTag("connectrpc.stream_type", spec.StreamType.String())

	data := map[string]interface{}{
		"procedure":   spec.Procedure,
		"service":     service,
		"method":      method,
		"stream_type": spec.StreamType.String(),
	}
	if ua := headers["User-Agent"]; len(ua) > 0 {
		data["user_agent"] = ua[0]
	}
	if ct := headers["Content-Type"]; len(ct) > 0 {
		data["content_type"] = ct[0]
	}

	scope.SetContext("ConnectRPC Request", data)
}

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

// splitProcedure splits a ConnectRPC procedure like "/package.Service/Method"
// into service and method parts.
func splitProcedure(procedure string) (service, method string) {
	procedure = strings.TrimPrefix(procedure, "/")
	if i := strings.LastIndex(procedure, "/"); i >= 0 {
		return procedure[:i], procedure[i+1:]
	}

	return procedure, ""
}
