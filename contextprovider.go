// Package contextprovider enables simpler context passing without requiring a change in function signature.
// It also enables localized and type-safe value passing ensuring access only to the functions needing them.
package contextprovider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

var runtimeCaller func(skip int) (pc uintptr, file string, line int, ok bool) = runtime.Caller

var contextMap = make(map[string]*context.Context)
var emptyContext = context.Background()
var emptyFunc = func() {}

// Provide provides a non-nil context to one or more receiver functions. It returns an error if the context
// is nil or if any of the passed functions are not actually functions.
func Provide(ctx context.Context, first any, rest ...any) error {
	if ctx == nil {
		return errors.New("provided ctx is nil")
	}
	var errs []error
	ctxPtr := &ctx
	funcs := make([]any, 1+len(rest))
	funcs[0] = first
	copy(funcs[1:], rest)

	for _, f := range funcs {
		key, err := getFuncKey(f)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		contextMap[key] = ctxPtr
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Inject should be called in the context-receiving function and it returns the context provided for it.
// If there's no context provided for the function, context.Background() will be returned and ok will be false.
// The free function should be used to clear the context if it is no longer needed.
func Inject() (ctx context.Context, ok bool, free func()) {
	return inject()
}

// InjectValue[T] should be called in the context-value-receiving function and it returns the value of type T for the given key.
// If there's no key and/or no value of type T associated with the key, the zero value of T will be returned and ok will be false.
// The free function should be used to clear the context only if there are no more values/context to be consumed.
func InjectValue[T any](key any) (value T, ok bool, free func()) {
	ctx, ok, free := inject()
	if !ok {
		return value, ok, free
	}
	actual, ok := ContextValue[T](ctx, key)
	return actual, ok, free
}

// ContextValue[T] is similar to InjectValue[T] except that you pass your own context
func ContextValue[T any](ctx context.Context, key any) (value T, ok bool) {
	actual, ok := ctx.Value(key).(T)
	if !ok {
		return value, ok
	}
	return actual, ok
}

// FreeContext frees the context provided for zero or more receiver functions. Usually you'd use the free functions returned
// by Inject / InjectValue[T] so you free it after consumption but this is sort of an escape-hatch in case
// you have a usecase where you want to free the context much later after consumption.
func FreeContext(funcs ...any) {
	for _, f := range funcs {
		key, _ := getFuncKey(f)
		freeFunc(key)
	}
}

func inject() (ctx context.Context, ok bool, free func()) {
	key := getInjectKey()
	if key == "" {
		return emptyContext, false, emptyFunc
	}
	ctxPtr, ok := contextMap[key]
	if !ok {
		return emptyContext, false, emptyFunc
	}
	return *ctxPtr, true, func() {
		freeFunc(key)
	}
}

func getInjectKey() string {
	pc, _, _, ok := runtimeCaller(3)
	if !ok {
		return ""
	}
	return getKey(pc)
}

func getFuncKey(f any) (string, error) {
	v := reflect.ValueOf(f)
	if v.Kind().String() != "func" {
		return "", fmt.Errorf("f %#v is not a function", f)
	}
	return getKey(v.Pointer()), nil
}

func getKey(pc uintptr) string {
	return runtime.FuncForPC(pc).Name()
}

func freeFunc(key string) {
	delete(contextMap, key)
}
