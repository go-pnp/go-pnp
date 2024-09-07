package fxutil

import (
	"go.uber.org/fx"
)

type OptionsBuilder struct {
	options         []fx.Option
	PrivateProvides bool
}

func (m *OptionsBuilder) InvokeIf(condition bool, fns ...any) {
	if condition {
		m.options = append(m.options, fx.Invoke(fns...))
	}
}

func (m *OptionsBuilder) Invoke(fns ...any) {
	m.options = append(m.options, fx.Invoke(fns...))
}

func (m *OptionsBuilder) ProvideIf(condition bool, fns ...any) {
	if !condition {
		return
	}
	if m.PrivateProvides {
		fns = append([]any{}, fns...)
		fns = append(fns, fx.Private)
	}

	m.options = append(m.options, fx.Provide(fns...))
}

func (m *OptionsBuilder) Supply(fns ...any) {
	if m.PrivateProvides {
		fns = append([]any{}, fns...)
		fns = append(fns, fx.Private)
	}
	m.options = append(m.options, fx.Supply(fns...))
}

func (m *OptionsBuilder) SupplyIf(condition bool, fns ...any) {
	if condition {
		m.Supply(fns...)
	}
}

func (m *OptionsBuilder) GroupSupply(group string, value any) {
	supplyArgs := []any{
		fx.Annotate(value, fx.ResultTags(`group:"`+group+`"`)),
	}
	if m.PrivateProvides {
		supplyArgs = append(supplyArgs, fx.Private)
	}

	m.options = append(m.options, fx.Supply(supplyArgs...))
}

func (m *OptionsBuilder) Provide(fns ...any) {
	if m.PrivateProvides {
		fns = append([]any{}, fns...)
		fns = append(fns, fx.Private)
	}
	m.options = append(m.options, fx.Provide(fns...))
}

func (m *OptionsBuilder) PublicProvide(fns ...any) {
	m.options = append(m.options, fx.Provide(fns...))
}

func (m *OptionsBuilder) PublicProvideIf(condition bool, fns ...any) {
	if condition {
		m.options = append(m.options, fx.Provide(fns...))
	}
}

func (m *OptionsBuilder) Option(opts ...fx.Option) {
	m.options = append(m.options, opts...)
}

func (m *OptionsBuilder) Decorate(fns ...any) {
	m.options = append(m.options, fx.Decorate(fns...))
}

func (m *OptionsBuilder) Build() fx.Option {
	return fx.Options(
		m.options...,
	)
}
