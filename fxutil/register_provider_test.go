package fxutil

import (
	"go.uber.org/fx"
	"testing"
)

type SomeStruct struct{ Val string }
type SomeInterface interface {
	isSomeInterface()
}
type SomeInterfaceImpl struct {
	Val string
}

func (SomeInterfaceImpl) isSomeInterface() {}

type SomeFunc func() string

type AllGroupsInput struct {
	fx.In
	SomeStructs     []SomeStruct    `group:"structs_group"`
	SomeStructsPtrs []*SomeStruct   `group:"struct_ptrs_group"`
	SomeInterfaces  []SomeInterface `group:"interfaces_group"`
	SomeFuncs       []SomeFunc      `group:"funcs_group"`
}

func TestNewGroupProvider(t *testing.T) {
	fx.New(
		fx.Provide(ProvideToGroup[SomeStruct]("structs_group", func() SomeStruct {
			return SomeStruct{Val: "hello"}
		})),
		fx.Provide(ProvideToGroup[*SomeStruct]("struct_ptrs_group", func() *SomeStruct {
			return &SomeStruct{Val: "hello1"}
		})),
		fx.Provide(ProvideToGroup[SomeInterface]("interfaces_group", func() SomeInterface {
			return SomeInterfaceImpl{Val: "world"}
		})),
		fx.Provide(ProvideToGroup[SomeFunc]("funcs_group", func() SomeFunc {
			return func() string {
				return "!"
			}
		})),
		fx.Invoke(func(inp AllGroupsInput) {
			if len(inp.SomeStructs) != 1 {
				t.Fatalf("expected 1 SomeStruct, got %d", len(inp.SomeStructs))
			}
			if inp.SomeStructs[0].Val != "hello" {
				t.Fatalf("expected SomeStruct.Val == 'hello', got %s", inp.SomeStructs[0].Val)
			}
			if len(inp.SomeStructsPtrs) != 1 {
				t.Fatalf("expected 1 *SomeStruct, got %d", len(inp.SomeStructsPtrs))
			}
			if inp.SomeStructsPtrs[0].Val != "hello1" {
				t.Fatalf("expected *SomeStruct.Val == 'hello', got %s", inp.SomeStructsPtrs[0].Val)
			}
			if len(inp.SomeInterfaces) != 1 {
				t.Fatalf("expected 1 SomeInterface, got %d", len(inp.SomeInterfaces))
			}
			if inp.SomeInterfaces[0].(SomeInterfaceImpl).Val != "world" {
				t.Fatalf("expected SomeInterfaceImpl.Val == 'world', got %s", inp.SomeInterfaces[0].(SomeInterfaceImpl).Val)
			}
			if len(inp.SomeFuncs) != 1 {
				t.Fatalf("expected 1 SomeFunc, got %d", len(inp.SomeFuncs))
			}
			if inp.SomeFuncs[0]() != "!" {
				t.Fatalf("expected SomeFunc() == '!', got %s", inp.SomeFuncs[0]())
			}
		}),
	)
}
