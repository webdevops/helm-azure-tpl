package base

import (
	"fmt"
	"reflect"
)

// deprecate wraps any function, logs a warning when it is called,
// and forwards the arguments and return values.
func (e *BaseEngine) Deprecate(name string, fn interface{}) interface{} {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("deprecate expects a function")
	}

	t := v.Type()

	// create a new function with the same signature as the original
	wrapper := reflect.MakeFunc(v.Type(), func(args []reflect.Value) []reflect.Value {
		e.Logger.Warn(fmt.Sprintf(`template function "%v" is deprecated and will be removed soon`, name))
		e.HandleCicdWarning(fmt.Errorf(`template function "%v" is deprecated and will be removed soon`, name))
		if t.IsVariadic() {
			return v.CallSlice(args)
		}
		return v.Call(args)
	})

	return wrapper.Interface()
}
