# Go PnP Application Framework (WIP)

Go PnP (Plug and Play) Framework is a simple and easy-to-use application framework for the Go programming language. It
provides a set of commonly-used libraries that are wrapped with Uber FX, making it easy for developers to create and run
their applications.

### [Uber Fx](https://github.com/uber-go/fx)

While Uber FX may not be everyone's cup of tea, it's worth giving it a chance and seeing how it can simplify your
development workflow and improve your code quality. By using Go PnP Framework with Uber FX, you can take advantage of
its many benefits without sacrificing convenience or flexibility.

## Motivation

Developing and deploying applications in Go can be challenging, especially when it comes to managing dependencies,
configuring different services, and handling cross-cutting concerns such as logging, metrics, and tracing. Go PnP (Plug
and Play) Framework aims to simplify these tasks by providing a set of commonly-used libraries that are wrapped with
Uber FX.

With Go PnP Framework, you no longer need to spend time writing verbose boilerplate code for each service, or worry
about managing dependencies and configuration. The framework takes care of all of that for you, ensuring that your
application is consistent across different services and deployments.

Moreover, Go PnP Framework helps with deployment by ensuring that configuration is the same across different services.
This means that you can easily deploy your application to different environments, knowing that it will behave
consistently regardless of the deployment environment.

Whether you're building a small microservice or a large-scale distributed system, Go PnP Framework can help you
accelerate development, improve code maintainability, and simplify deployment. So why not give it a try and see how it
can simplify your life as a Go developer?

After I fully implement the framework, it will be possible to write fully functional applications with just a few lines
of code like this:

```go
func main(){
    app := fx.New(
        pnpzap.Module(),
        pnpzapsentry.Module(),
		
        pnptracing.Module(),

        pnpprometheus.Module(),
        
        pnpgrpcserver.Module(),
        pnpgrpcserverlog.Module(),
        pnpgrpcservermetrics.Module(),
        pnpgrpcservertracing.Module(),
        
        pnpsarama.Module(),
        pnpsarama.ConsumerModule(),
        // Your business logic modules.
        // And that's it! You have a fully functional application without boilerplate code.
    )
    app.Run()
}
```

## Concepts Used in Framework

Go PnP Framework uses the following concepts:

- **Dependency Injection:** The framework uses Uber FX for dependency injection, which makes it easy to manage
  dependencies and improve code maintainability.
- **Modularity:** The framework is designed to be modular, allowing developers to add or remove functionality as needed.
- **Extendability** Each module in the framework is designed to be extendable, allowing developers to add their own
  functionality to the framework.

## List of Modules
- [x] [Healthcheck](https://github.com/go-pnp/go-pnp/tree/master/healthcheck/pnphealthcheck)
- [x] [HTTP Server](https://github.com/go-pnp/go-pnp/tree/master/http/pnphttpserver)
  - [ ] HTTP middleware for logging, metrics aggregation, and tracing
- [x] [Fiber HTTP Server](https://github.com/go-pnp/go-pnp/tree/master/http/pnpfiber)
  - Endpoints
    - [x] [Healthcheck](https://github.com/go-pnp/go-pnp/tree/master/http/pnpfiberhealthcheck)
    - [x] [Prometheus](https://github.com/go-pnp/go-pnp/tree/master/http/pnpfiberprometheus)
  - Middleware
    - [ ] Fiber opentelemetry
- [x] [Zap logging](https://github.com/go-pnp/go-pnp/tree/master/logging/pnpzap)
  - [ ] Zap hooks for sentry
- [ ] Logrus logging
  - [ ] Logrus hooks for sentry
- [x] [gRPC Server](https://github.com/go-pnp/go-pnp/tree/master/grpc/pnpgrpcserver)
  - [ ] gRPC interceptors for logging, metrics aggregation, and tracing
- [x] [gRPC Client](https://github.com/go-pnp/go-pnp/tree/master/grpc/pnpgrpcclient)
- [x] [gRPC Web](https://github.com/go-pnp/go-pnp/tree/master/grpc/pnpgrpcweb)
- [x] [SQL](https://github.com/go-pnp/go-pnp/tree/master/sql/pnpsql)
- [x] [SQLx](https://github.com/go-pnp/go-pnp/tree/master/sql/pnpsqlx)
- [x] [Gorm](https://github.com/go-pnp/go-pnp/tree/master/sql/pnpgorm)
- [x] [Pgx](https://github.com/go-pnp/go-pnp/tree/master/sql/pnppgx)
- [ ] Nats
  - [x] [JetStream Subscriptions](https://github.com/go-pnp/go-pnp/tree/master/nats/pnpnats)
  - [ ] Regular Subscriptions

- [x] [Prometheus metrics server](https://github.com/go-pnp/go-pnp/tree/master/prometheus/pnpprometheus)
- [x] [Redis](https://github.com/go-pnp/go-pnp/tree/master/pnpredis)
- [ ] Sarama client
- [ ] TODO: Add more modules

## Usage Examples

Each module contains a example_test.go file that demonstrates how to use the module.

Here's a simple example of pnphttpserver module usage:

```go
type Handler struct {
}

func (h Handler) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("World1"))
}

func (h Handler) RegisterEndpoints(mux *mux.Router) {
	mux.Path("/hello").HandlerFunc(h.Hello)
}

func TestApp(t *testing.T) {
	fxutil.StartApp(
		Module(),
		fx.Supply(Handler{}),

		// Register our application endpoints
		ProvideMuxHandlerRegistrar(func(handler Handler) MuxHandlerRegistrar {
			return handler.RegisterEndpoints
		}),

		// Register middleware
		ProvideMuxMiddlewareFunc(func() mux.MiddlewareFunc {
			return func(mux http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello from middleware\n"))

					mux.ServeHTTP(w, r)
				})
			}
		}),
	)
}

```

## Contribution Guide

We welcome contributions to Go PnP Framework! To contribute, please follow these steps:

- Fork the repository.
- Create a new branch for your changes.
- Make your changes and commit them to your branch.
- Submit a pull request.

Please be consistent with the existing code style and with existing module structure.
