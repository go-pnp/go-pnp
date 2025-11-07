package fxutil

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"go.uber.org/fx"
)

// GroupProvider returns an fx.Annotated and checks that the target is a function that returns at least one value of type T.
// In most cases, you don't need it unless you writing your own module.
func GroupProvider[T any](groupName string, target any) any {
	if groupName == "" {
		panic(errors.New("empty group name"))
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Func {
		panic(errors.New("target should be function"))
	}

	zeroT := new(T)
	requiredType := reflect.TypeOf(zeroT).Elem()
	annotations := []fx.Annotation{
		fx.ResultTags(fmt.Sprintf(`group:"%s"`, groupName)),
	}
	var foundRequiredType bool
	for i := range targetType.NumOut() {
		outType := targetType.Out(i)
		if outType == requiredType {
			foundRequiredType = true
			break
		}
		if requiredType.Kind() == reflect.Interface && outType.Implements(requiredType) {
			foundRequiredType = true
			annotations = append(annotations, fx.As(zeroT))
			break
		}
	}

	if !foundRequiredType {
		panic(errors.Errorf("provider should return at least one value of type %v", *zeroT))
	}

	return fx.Annotate(
		target,
		annotations...,
	)
}
