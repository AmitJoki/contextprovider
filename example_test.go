package contextprovider_test

import (
	"context"
	"fmt"
	"time"

	"github.com/AmitJoki/contextprovider"
)

type userKey string

var loggedInKey userKey = "loggedInAt"
var value = true
var timeoutInSeconds = 3

func ExampleProvide() {
	ctx := context.WithValue(context.Background(), loggedInKey, value)
	contextprovider.Provide(ctx, ExampleInjectValue)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutInSeconds)*time.Second)
	context.AfterFunc(ctx, cancel) // to satisfy https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/lostcancel
	contextprovider.Provide(ctx, ExampleInject)
}

func ExampleInjectValue() {
	loggedIn, _, free := contextprovider.InjectValue[bool](loggedInKey)
	fmt.Printf("Logged in: %v", loggedIn)
	free()
	// Output:
	// Logged in: true
}

func ExampleInject() {
	ctx, _, free := contextprovider.Inject()
	<-ctx.Done()
	fmt.Printf("Context cancelled after %v seconds", timeoutInSeconds)
	free()
	// Output:
	// Context cancelled after 3 seconds
}

func init() {
	ExampleProvide()
}
