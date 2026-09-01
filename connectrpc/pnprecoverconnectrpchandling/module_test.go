package pnprecoverconnectrpchandling

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestInterceptor_WrapUnary_UsesPanicHandler(t *testing.T) {
	handlerError := errors.New("panic handled")

	var gotPanicValue any

	interceptor := NewInterceptor(newOptions(WithPanicHandler(
		func(_ context.Context, _ connect.Spec, panicValue any) error {
			gotPanicValue = panicValue

			return handlerError
		},
	))).Value

	unaryFunc := interceptor.WrapUnary(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		panic("boom")
	})

	_, err := unaryFunc(context.Background(), connect.NewRequest(&emptypb.Empty{}))

	require.ErrorIs(t, err, handlerError)
	require.Equal(t, "boom", gotPanicValue)
}
