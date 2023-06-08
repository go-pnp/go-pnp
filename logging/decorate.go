package logging

import "go.uber.org/fx"

type DecorateNamedParams struct {
	fx.In
	Logger *Logger `optional:"true"`
}

func DecorateNamed(name string) fx.Option {
	return fx.Decorate(func(params DecorateNamedParams) *Logger {
		if params.Logger != nil {
			return nil
		}

		return params.Logger.Named(name)
	})
}
