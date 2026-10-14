package pnpwatermillrecovery_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-pnp/go-pnp/pkg/ordering"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/go-pnp/go-pnp/watermill/pnpwatermillrecovery"
)

type handlerMiddlewares struct {
	fx.In

	Items []ordering.OrderedItem[message.HandlerMiddleware] `group:"pnpwatermill.handler_middlewares"`
}

func startMiddleware(t *testing.T, options ...fx.Option) ordering.OrderedItem[message.HandlerMiddleware] {
	t.Helper()

	var middlewares handlerMiddlewares

	app := fxtest.New(t, append(options, fx.Populate(&middlewares))...)
	app.RequireStart()
	t.Cleanup(app.RequireStop)

	require.Len(t, middlewares.Items, 1)

	return middlewares.Items[0]
}

func panickingHandler(*message.Message) ([]*message.Message, error) {
	panic("boom")
}

func TestPanicIsReturnedAsErrorWithStackOfPanickingHandler(t *testing.T) {
	middleware := startMiddleware(t, pnpwatermillrecovery.Module()).Value

	produced, err := middleware(panickingHandler)(message.NewMessage("id", nil))

	assert.Nil(t, produced)
	require.EqualError(t, err, "panic: boom")
	assert.Contains(t, fmt.Sprintf("%+v", err), "pnpwatermillrecovery_test.panickingHandler")
}

func TestHandlerResultIsPassedThrough(t *testing.T) {
	middleware := startMiddleware(t, pnpwatermillrecovery.Module()).Value
	handlerError := errors.New("handler failed")
	handlerMessages := []*message.Message{message.NewMessage("produced", nil)}

	produced, err := middleware(func(*message.Message) ([]*message.Message, error) {
		return handlerMessages, handlerError
	})(message.NewMessage("id", nil))

	assert.Equal(t, handlerMessages, produced)
	require.ErrorIs(t, err, handlerError)
}

func TestModuleUsesConfiguredOrder(t *testing.T) {
	middleware := startMiddleware(t, pnpwatermillrecovery.Module(pnpwatermillrecovery.WithOrder(7)))

	assert.Equal(t, 7, middleware.Order)
}
