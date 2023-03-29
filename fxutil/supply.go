package fxutil

import "go.uber.org/fx"

func PrivateSupply[T any](val T) fx.Option {
	return fx.Provide(func() T {
		return val
	}, fx.Private)
}

func PrivateSupply2[T1, T2 any](val1 T1, val2 T2) fx.Option {
	return fx.Provide(
		func() T1 {
			return val1
		},
		func() T2 {
			return val2
		},
		fx.Private)
}

func PrivateSupply3[T1, T2, T3 any](val1 T1, val2 T2, val3 T3) fx.Option {
	return fx.Provide(
		func() T1 {
			return val1
		},
		func() T2 {
			return val2
		},
		func() T3 {
			return val3
		},
		fx.Private)
}
