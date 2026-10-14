package pnphttpserverrecovery_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/go-pnp/go-pnp/http/pnphttpserverrecovery"
)

type requestContextKey struct{}

type recordedLogEntry struct {
	level               string
	message             string
	err                 error
	requestContextValue any
}

type recordingLogDelegate struct {
	entries *[]recordedLogEntry
	err     error
}

func (d recordingLogDelegate) Info(ctx context.Context, message string, _ ...interface{}) {
	d.record(ctx, "info", message)
}

func (d recordingLogDelegate) Warn(ctx context.Context, message string, _ ...interface{}) {
	d.record(ctx, "warn", message)
}

func (d recordingLogDelegate) Debug(ctx context.Context, message string, _ ...interface{}) {
	d.record(ctx, "debug", message)
}

func (d recordingLogDelegate) Error(ctx context.Context, message string, _ ...interface{}) {
	d.record(ctx, "error", message)
}

func (d recordingLogDelegate) WithFields(map[string]interface{}) logging.Delegate { return d }
func (d recordingLogDelegate) WithField(string, interface{}) logging.Delegate     { return d }
func (d recordingLogDelegate) Named(string) logging.Delegate                      { return d }
func (d recordingLogDelegate) SkipCallers(int) logging.Delegate                   { return d }

func (d recordingLogDelegate) WithError(err error) logging.Delegate {
	d.err = err

	return d
}

func (d recordingLogDelegate) record(ctx context.Context, level, message string) {
	*d.entries = append(*d.entries, recordedLogEntry{
		level:               level,
		message:             message,
		err:                 d.err,
		requestContextValue: ctx.Value(requestContextKey{}),
	})
}

func newRecordingLogger() (*logging.Logger, *[]recordedLogEntry) {
	var entries []recordedLogEntry

	return &logging.Logger{Delegate: recordingLogDelegate{entries: &entries}}, &entries
}

type handlerMiddlewares struct {
	fx.In

	Items []ordering.OrderedItem[pnphttpserver.HandlerMiddleware] `group:"pnp_http_server.handler_middlewares"`
}

func startMiddleware(t *testing.T, options ...fx.Option) pnphttpserver.HandlerMiddleware {
	t.Helper()

	var middlewares handlerMiddlewares

	app := fxtest.New(t, append(options, fx.Populate(&middlewares))...)
	app.RequireStart()
	t.Cleanup(app.RequireStop)

	require.Len(t, middlewares.Items, 1)

	return middlewares.Items[0].Value
}

func serve(middleware pnphttpserver.HandlerMiddleware, handler http.HandlerFunc) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = request.WithContext(context.WithValue(request.Context(), requestContextKey{}, "request-context"))

	recorder := httptest.NewRecorder()
	middleware(handler).ServeHTTP(recorder, request)

	return recorder
}

func panickingHandler(http.ResponseWriter, *http.Request) {
	panic("boom")
}

func abortingHandler(http.ResponseWriter, *http.Request) {
	panic(http.ErrAbortHandler)
}

func teapotPanicHandler(w http.ResponseWriter, _ any) {
	w.WriteHeader(http.StatusTeapot)
}

type recordingPanicHandler struct {
	handledPanicValues []any
}

func (h *recordingPanicHandler) handle(w http.ResponseWriter, panicValue any) {
	h.handledPanicValues = append(h.handledPanicValues, panicValue)

	teapotPanicHandler(w, panicValue)
}

func TestPanicIsLoggedWithStackAndRequestContextBeforeDefaultResponse(t *testing.T) {
	logger, entries := newRecordingLogger()

	recorder := serve(startMiddleware(t, pnphttpserverrecovery.Module(), fx.Supply(logger)), panickingHandler)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "Internal Server Error\n", recorder.Body.String())
	require.Len(t, *entries, 1)
	assert.Equal(t, "error", (*entries)[0].level)
	require.EqualError(t, (*entries)[0].err, "panic: boom")
	assert.Contains(t, fmt.Sprintf("%+v", (*entries)[0].err), "pnphttpserverrecovery_test.panickingHandler")
	assert.Equal(t, "request-context", (*entries)[0].requestContextValue)
}

func TestPanicIsAnsweredWithoutLogger(t *testing.T) {
	recorder := serve(startMiddleware(t, pnphttpserverrecovery.Module()), panickingHandler)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestCustomPanicHandlerIsCalledAfterPanicIsLogged(t *testing.T) {
	logger, entries := newRecordingLogger()
	panicHandler := &recordingPanicHandler{}

	recorder := serve(startMiddleware(t,
		pnphttpserverrecovery.Module(pnphttpserverrecovery.WithPanicHandler(panicHandler.handle)),
		fx.Supply(logger),
	), panickingHandler)

	assert.Equal(t, http.StatusTeapot, recorder.Code)
	assert.Equal(t, []any{"boom"}, panicHandler.handledPanicValues)
	require.Len(t, *entries, 1)
	require.EqualError(t, (*entries)[0].err, "panic: boom")
}

func TestAbortHandlerPanicIsPassedToPanicHandlerWithoutLogging(t *testing.T) {
	logger, entries := newRecordingLogger()
	panicHandler := &recordingPanicHandler{}

	recorder := serve(startMiddleware(t,
		pnphttpserverrecovery.Module(pnphttpserverrecovery.WithPanicHandler(panicHandler.handle)),
		fx.Supply(logger),
	), abortingHandler)

	assert.Equal(t, http.StatusTeapot, recorder.Code)
	assert.Equal(t, []any{http.ErrAbortHandler}, panicHandler.handledPanicValues)
	assert.Empty(t, *entries)
}

func TestRequestWithoutPanicIsPassedThrough(t *testing.T) {
	logger, entries := newRecordingLogger()
	panicHandler := &recordingPanicHandler{}

	recorder := serve(startMiddleware(t,
		pnphttpserverrecovery.Module(pnphttpserverrecovery.WithPanicHandler(panicHandler.handle)),
		fx.Supply(logger),
	), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "created", recorder.Body.String())
	assert.Empty(t, panicHandler.handledPanicValues)
	assert.Empty(t, *entries)
}

func TestModuleUsesPanicHandlerFromContainer(t *testing.T) {
	recorder := serve(startMiddleware(t,
		pnphttpserverrecovery.Module(pnphttpserverrecovery.WithPanicHandlerFromContainer()),
		fx.Supply(pnphttpserverrecovery.PanicHandler(teapotPanicHandler)),
	), panickingHandler)

	assert.Equal(t, http.StatusTeapot, recorder.Code)
}

func TestModuleProvidesDefaultPanicHandlerToContainer(t *testing.T) {
	var panicHandler pnphttpserverrecovery.PanicHandler

	app := fxtest.New(t, pnphttpserverrecovery.Module(), fx.Populate(&panicHandler))
	app.RequireStart()
	t.Cleanup(app.RequireStop)

	recorder := httptest.NewRecorder()
	panicHandler(recorder, "boom")

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestPrivateModuleKeepsItsOwnPanicHandlerNextToACustomOne(t *testing.T) {
	var defaultServerMiddlewares, customServerMiddlewares handlerMiddlewares

	app := fxtest.New(t,
		fx.Module("default-server",
			pnphttpserverrecovery.Module(pnphttpserverrecovery.WithFxPrivate()),
			fx.Invoke(func(middlewares handlerMiddlewares) { defaultServerMiddlewares = middlewares }),
		),
		fx.Module("custom-server",
			pnphttpserverrecovery.Module(pnphttpserverrecovery.WithFxPrivate(), pnphttpserverrecovery.WithPanicHandler(teapotPanicHandler)),
			fx.Invoke(func(middlewares handlerMiddlewares) { customServerMiddlewares = middlewares }),
		),
	)
	app.RequireStart()
	t.Cleanup(app.RequireStop)

	require.Len(t, defaultServerMiddlewares.Items, 1)
	require.Len(t, customServerMiddlewares.Items, 1)
	assert.Equal(t, http.StatusInternalServerError, serve(defaultServerMiddlewares.Items[0].Value, panickingHandler).Code)
	assert.Equal(t, http.StatusTeapot, serve(customServerMiddlewares.Items[0].Value, panickingHandler).Code)
}

func TestPanicHandlerFromContainerMustBeProvided(t *testing.T) {
	app := fx.New(
		pnphttpserverrecovery.Module(pnphttpserverrecovery.WithPanicHandlerFromContainer()),
		fx.Invoke(func(handlerMiddlewares) {}),
		fx.NopLogger,
	)

	require.Error(t, app.Err())
}
