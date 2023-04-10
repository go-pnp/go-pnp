package fxutil

import (
	"go.uber.org/fx"
)

type OptionsBuilder struct {
	options         []fx.Option
	PrivateProvides bool
}

func (m *OptionsBuilder) InvokeIf(condition bool, fns ...interface{}) {
	if condition {
		m.options = append(m.options, fx.Invoke(fns...))
	}
}

func (m *OptionsBuilder) Invoke(fns ...interface{}) {
	m.options = append(m.options, fx.Invoke(fns...))
}

func (m *OptionsBuilder) ProvideIf(condition bool, fns ...interface{}) {
	if !condition {
		return
	}
	if m.PrivateProvides {
		fns = append([]interface{}{}, fns...)
		fns = append(fns, fx.Private)
	}

	m.options = append(m.options, fx.Provide(fns...))
}

// OptionsBuilderSupply is a helper to supply a value using Provide. It's required to supply a value privately.
func OptionsBuilderSupply[T any](m *OptionsBuilder, val T) {
	m.Provide(func() T {
		return val
	})
}

// OptionsBuilderGroupSupply is a helper to supply a value using Provide. It's required to supply a value privately.
func OptionsBuilderGroupSupply[T any](m *OptionsBuilder, group string, val T) {
	m.Provide(fx.Annotated{
		Group:  group,
		Target: func() T { return val },
	})
}

func (m *OptionsBuilder) Provide(fns ...interface{}) {
	if m.PrivateProvides {
		fns = append([]interface{}{}, fns...)
		fns = append(fns, fx.Private)
	}
	m.options = append(m.options, fx.Provide(fns...))
}

func (m *OptionsBuilder) Option(opts ...fx.Option) {
	m.options = append(m.options, opts...)
}

// We don't use Supply as it do not support fx.PrivateProvides yet

func (m *OptionsBuilder) Build() fx.Option {
	return fx.Options(
		m.options...,
	)
}
