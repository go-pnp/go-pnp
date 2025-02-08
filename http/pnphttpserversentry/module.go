package pnphttpserversentry

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"go.uber.org/fx"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.Supply(options)
	moduleBuilder.Provide(pnphttpserver.HandlerMiddlewareProvider(newMiddleware))

	return moduleBuilder.Build()
}

func newMiddleware(options *options, client *sentry.Client) ordering.OrderedItem[pnphttpserver.HandlerMiddleware] {
	return ordering.OrderedItem[pnphttpserver.HandlerMiddleware]{
		Value: func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				hub := sentry.GetHubFromContext(request.Context())
				var scope *sentry.Scope
				if hub == nil {
					scope = sentry.NewScope()
					hub = sentry.NewHub(client, scope)
					request = request.WithContext(sentry.SetHubOnContext(request.Context(), hub))
				} else {
					scope = hub.PushScope()
					defer hub.PopScope()
				}

				span := sentry.StartSpan(request.Context(), fmt.Sprintf("%s %s", request.Method, request.URL.Path))
				defer span.Finish()
				request = request.WithContext(span.Context())

				scope.SetSpan(span)
				scope.SetRequest(request)

				wrapHandler(handler, client, "nested smth").ServeHTTP(writer, request)
			})
		},
		Order: options.order,
	}
}

func wrapHandler(handler http.Handler, sentryClient *sentry.Client, op string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := sentry.GetHubFromContext(r.Context())
		var scope *sentry.Scope
		if hub == nil {
			scope = sentry.NewScope()
			hub = sentry.NewHub(sentryClient, scope)
			r = r.WithContext(sentry.SetHubOnContext(r.Context(), hub))
		} else {
			scope = hub.PushScope()
			defer hub.PopScope()
		}
		span := sentry.StartSpan(r.Context(), op)
		defer span.Finish()

		r = r.WithContext(span.Context())

		scope.SetSpan(span)

		handler.ServeHTTP(w, r)
	})
}
