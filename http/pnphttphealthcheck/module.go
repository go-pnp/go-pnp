package pnphttphealthcheck

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-pnp/go-pnp/healthcheck/pnphealthcheck"
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}

	moduleBuilder.Supply(options)
	moduleBuilder.Provide(NewHealthcheckHandler)
	moduleBuilder.Provide(pnphttpserver.MuxHandlerRegistrarProvider(NewMuxHandlerRegistrar))

	return moduleBuilder.Build()
}

func WriteResponse(alive bool, checks map[string]error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	if !alive {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	errInfo := make(map[string]any, len(checks))
	for name, err := range checks {
		multiErr, ok := err.(interface {
			Unwrap() []error
		})
		if ok {
			errStrings := make([]string, 0, len(multiErr.Unwrap()))
			for _, err := range multiErr.Unwrap() {
				errStrings = append(errStrings, err.Error())
			}
			errInfo[name] = errStrings
		} else {
			errInfo[name] = err.Error()
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"alive":       alive,
		"checkErrors": errInfo,
	})
}

type HealthCheckHandler http.HandlerFunc

func NewHealthcheckHandler(
	options *options,
	healthResolver *pnphealthcheck.HealthResolver,
) HealthCheckHandler {
	return func(writer http.ResponseWriter, request *http.Request) {
		checkResults, alive := healthResolver.Resolve(request.Context())
		if options.responseWriter != nil {
			options.responseWriter(alive, checkResults, writer)

			return
		}

		if alive {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

type NewMuxHandlerRegistrarParams struct {
	fx.In
	Options *options
	Handler HealthCheckHandler
	Logger  *logging.Logger `optional:"true"`
}

func NewMuxHandlerRegistrar(params NewMuxHandlerRegistrarParams) pnphttpserver.MuxHandlerRegistrar {
	return pnphttpserver.MuxHandlerRegistrarFunc(func(mux *mux.Router) {
		params.Logger.Named("http-healthchecks").Debug(context.Background(), "Registering healthcheck handler")
		mux.Methods(params.Options.method).Path(params.Options.path).HandlerFunc(params.Handler)
	})
}
