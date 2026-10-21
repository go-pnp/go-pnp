package pnpzapsentry_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-pnp/go-pnp/logging/pnpzapsentry"
)

func TestZapCoreWriteCapturesEventBeforeReturning(t *testing.T) {
	logger, client, capturedEvents := newSentryLogger(t)

	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ me { id } }"}`))
	hub := sentry.NewHub(client, sentry.NewScope())
	hub.Scope().SetRequest(request)
	ctx := sentry.SetHubOnContext(context.Background(), hub)

	logger.With(zap.Field{Key: "context", Type: zapcore.SkipType, Interface: ctx}).Error("middleware failed before the body was read")

	if _, err := io.ReadAll(request.Body); err != nil {
		t.Fatal(err)
	}

	event := receiveEvent(t, capturedEvents)
	if event.Request == nil || event.Request.Data != "" {
		t.Fatalf("expected event with empty request data captured before the body was read, got %+v", event.Request)
	}
}

func TestZapCoreWriteAttachesCallerStacktrace(t *testing.T) {
	const testFunction = "TestZapCoreWriteAttachesCallerStacktrace"

	logger, _, capturedEvents := newSentryLogger(t)

	logger.Error("message")

	messageEvent := receiveEvent(t, capturedEvents)
	if len(messageEvent.Threads) != 1 {
		t.Fatalf("expected one thread, got %d", len(messageEvent.Threads))
	}

	assertLastFrameFunction(t, messageEvent.Threads[0].Stacktrace, testFunction)

	logger.With(zap.Error(errors.New("error without stack"))).Error("error without stack")

	assertLastFrameFunction(t, lastException(t, receiveEvent(t, capturedEvents)).Stacktrace, testFunction)

	logger.With(zap.Error(newErrorWithStack())).Error("error with stack")

	assertLastFrameFunction(t, lastException(t, receiveEvent(t, capturedEvents)).Stacktrace, "newErrorWithStack")
}

func newSentryLogger(t *testing.T) (*zap.Logger, *sentry.Client, <-chan *sentry.Event) {
	t.Helper()

	capturedEvents := make(chan *sentry.Event, 1)

	client, err := sentry.NewClient(sentry.ClientOptions{
		BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			capturedEvents <- event

			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	sentryOption, err := pnpzapsentry.NewZapCoreWrapperOption(pnpzapsentry.NewZapCoreWrapperOptionParams{
		SentryClient: client,
		Config:       &pnpzapsentry.Config{FlushTimeout: time.Second, ReportLevel: sentry.LevelError},
	})
	if err != nil {
		t.Fatal(err)
	}

	underlying := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.DebugLevel)

	return zap.New(underlying, sentryOption, zap.AddCaller()), client, capturedEvents
}

func receiveEvent(t *testing.T, capturedEvents <-chan *sentry.Event) *sentry.Event {
	t.Helper()

	select {
	case event := <-capturedEvents:
		return event
	case <-time.After(time.Second):
		t.Fatal("event was not captured")

		return nil
	}
}

func lastException(t *testing.T, event *sentry.Event) sentry.Exception {
	t.Helper()

	if len(event.Exception) == 0 {
		t.Fatal("expected exception event")
	}

	return event.Exception[len(event.Exception)-1]
}

func assertLastFrameFunction(t *testing.T, stacktrace *sentry.Stacktrace, function string) {
	t.Helper()

	if stacktrace == nil || len(stacktrace.Frames) == 0 {
		t.Fatal("expected stacktrace")
	}

	if lastFrame := stacktrace.Frames[len(stacktrace.Frames)-1]; lastFrame.Function != function {
		t.Fatalf("expected last frame %s, got %s (%s:%d)", function, lastFrame.Function, lastFrame.AbsPath, lastFrame.Lineno)
	}
}

func newErrorWithStack() error {
	return pkgerrors.New("error with stack")
}
