package contextprovider_test

import (
	"context"
	"testing"

	"github.com/AmitJoki/contextprovider"
)

type userDefinedType string

var question userDefinedType = "The Ultimate Question of Life, the Universe and Everything"
var answer int64 = 42

var backgroundContext = context.Background()

func TestRuntimeCaller(t *testing.T) {
	restore := contextprovider.StubRuntimeCaller()
	_, ok, _ := contextprovider.Inject()
	if ok {
		t.Error("Invalid caller information returns ok = true")
	}
	restore()
}

func TestInjectValue(t *testing.T) {
	var provider, intermediate, receiver func()
	provider = func() {
		ctx := context.WithValue(backgroundContext, question, answer)
		contextprovider.Provide(ctx, receiver)
		intermediate()
	}
	intermediate = func() {
		v, ok, _ := contextprovider.InjectValue[int64](question)
		if v == answer || ok {
			t.Errorf("Value %v of type %T accessible outside intended receiver function", v, v)
		}
		receiver()
	}
	receiver = func() {
		v, ok, free := contextprovider.InjectValue[int64](question)
		if v != answer || !ok {
			t.Errorf("Value %v of type %T inaccessible inside intended receiver function", v, v)
		}
		free()
	}
	provider()
}

func TestFreeingContext(t *testing.T) {
	var provider, receiver func()
	provider = func() {
		ctx := context.WithValue(backgroundContext, question, answer)
		contextprovider.Provide(ctx, receiver)
	}
	t.Run("free from InjectValue", func(t *testing.T) {
		receiver = func() {
			v, ok, free := contextprovider.InjectValue[int64](question)
			if v != answer || !ok {
				t.Errorf("Value %v of type %T inaccessible inside intended receiver function", v, v)
			}
			free()
			v, ok, _ = contextprovider.InjectValue[int64](question)
			if v == answer || ok {
				t.Errorf("Value %v of type %T persists even after freeing the context", v, v)
			}
		}
		provider()
		receiver()
	})
	t.Run("free from Inject", func(t *testing.T) {
		receiver = func() {
			ctx, _, free := contextprovider.Inject()
			if ctx == backgroundContext {
				t.Errorf("ctx is empty before calling free")
			}
			free()
			ctx, _, free = contextprovider.Inject()
			if ctx != backgroundContext {
				t.Errorf("ctx is not empty after calling free")
			}
			free()
		}
		provider()
		receiver()
	})

	t.Run("explicit free", func(t *testing.T) {
		receiver = func() {
			ctx, _, _ := contextprovider.Inject()
			if ctx == backgroundContext {
				t.Errorf("ctx is empty before calling free")
			}
			contextprovider.FreeContext(receiver)
			ctx, _, _ = contextprovider.Inject()
			if ctx != backgroundContext {
				t.Errorf("ctx is not empty after calling free")
			}
		}
		provider()
		receiver()
	})
}

func TestProvideForMultipleFuncs(t *testing.T) {
	var provider, receiver1, receiver2 func()
	provider = func() {
		ctx := context.WithValue(backgroundContext, question, answer)
		contextprovider.Provide(ctx, receiver1, receiver2)
	}
	receiver1 = func() {
		v, ok, free := contextprovider.InjectValue[int64](question)
		if v != answer || !ok {
			t.Error("receiver1 did not receive the context")
		}
		free()
	}
	receiver2 = func() {
		v, ok, free := contextprovider.InjectValue[int64](question)
		if v != answer || !ok {
			t.Error("receiver1 did not receive the context")
		}
		free()
	}
	provider()
	receiver1()
	receiver2()
}

func TestProvide(t *testing.T) {
	err := contextprovider.Provide(backgroundContext, 2)
	if err == nil {
		t.Error("Provide accepts non-function receivers without an error")
	}
	//lint:ignore SA1012 It is passed only for testing purpose
	err = contextprovider.Provide(nil, TestInjectValue)
	if err == nil {
		t.Error("Provide accepts nil context without an error")
	}
}

func TestContextValue(t *testing.T) {
	ctx := context.WithValue(backgroundContext, question, answer)
	v, ok := contextprovider.ContextValue[int64](ctx, question)
	if v != answer || !ok {
		t.Error("ContextValue does not return the right value for the correct type and key")
	}
	_, ok = contextprovider.ContextValue[string](ctx, question)
	if ok {
		t.Error("ContextValue returns ok = true for a non matching key/value type")
	}
}
