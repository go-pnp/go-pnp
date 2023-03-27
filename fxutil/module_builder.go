package fxutil

import "go.uber.org/fx"

type ModuleBuilder struct {
	ModuleName string
	Options    []fx.Option
}

func (m *ModuleBuilder) InvokeIf(condition bool, fns ...interface{}) {
	if condition {
		m.Options = append(m.Options, fx.Invoke(fns...))
	}
}

func (m *ModuleBuilder) Invoke(fns ...interface{}) {
	m.Options = append(m.Options, fx.Invoke(fns...))
}

func (m *ModuleBuilder) ProvideIf(condition bool, fns ...interface{}) {
	if condition {
		m.Options = append(m.Options, fx.Provide(fns...))
	}
}

func (m *ModuleBuilder) Provide(fns ...interface{}) {
	m.Options = append(m.Options, fx.Provide(fns...))
}

func (m *ModuleBuilder) SupplyIf(condition bool, fns ...interface{}) {
	if condition {
		m.Options = append(m.Options, fx.Supply(fns...))
	}
}

func (m *ModuleBuilder) Supply(fns ...interface{}) {
	m.Options = append(m.Options, fx.Supply(fns...))
}

func (m *ModuleBuilder) Build() fx.Option {
	return fx.Module(
		m.ModuleName,
		m.Options...,
	)
}
