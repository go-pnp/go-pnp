package fxutil

import (
	"reflect"

	"github.com/pkg/errors"
	"go.uber.org/fx"
)

func GroupProvider[T any](groupName string, target interface{}) fx.Annotated {
	if groupName == "" {
		panic(errors.New("empty group name"))
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Func {
		panic(errors.New("target should be function"))
	}

	zeroT := new(T)
	requiredType := reflect.TypeOf(zeroT).Elem()

	var foundRequiredType bool
	for i := 0; i < targetType.NumOut(); i++ {
		outType := targetType.Out(i)
		if outType == requiredType {
			foundRequiredType = true
			break
		}
	}

	if !foundRequiredType {
		panic(errors.Errorf("provider should return at least one value of type %v", zeroT))
	}

	return fx.Annotated{
		Group:  groupName,
		Target: target,
	}
}
