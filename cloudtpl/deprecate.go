package cloudtpl

import (
	"fmt"
	"reflect"
)

// deprecate wraps any function, logs a warning when it is called,
// and forwards the arguments and return values.
func (e *AzureTemplateExecutor) deprecate(name string, fn interface{}) interface{} {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("deprecate expects a function")
	}

	t := v.Type()

	// Create a new function with the same signature as the original
	wrapper := reflect.MakeFunc(v.Type(), func(args []reflect.Value) []reflect.Value {
		e.logger.Warn(fmt.Sprintf(`template function "%v" is deprecated`, name))
		e.handleCicdWarning(fmt.Errorf(`template function "%v" is deprecated`, name))
		// Use CallSlice for variadic functions, Call for standard functions
		if t.IsVariadic() {
			return v.CallSlice(args)
		}
		return v.Call(args)
	})

	return wrapper.Interface()
}
