# contextprovider

Package contextprovider enables simpler context passing without requiring a change in function signature.
It also enables localized and type-safe value passing ensuring access only to the functions needing them.


## Motivation

I am learning Go and I came across this [blog post](https://faiface.github.io/post/context-should-go-away-go2/).

1. Contexts are pervasive and require changes to function signature.
2. Functions that do not care about a context still has to change its signature just so it can drill that context down another level.
3. There's boilerplate involved in getting type-safe data back.

I prefer learning by doing and thought to myself if the above pain points can be solved by a package and thus contextprovider was born. This is purely academic and may look like magic because it is. It uses `reflect` and `runtime` packages under the hood which adds a performance overhead. I do not have any production grade experience in writing Go so use the package at your own risk.

# Concepts

## Provider and Receiver

Provider provides the context. Assume there are 2 functions `foo()`, `bar()`. `foo` is a provider which provides the context that is needed by `bar`.

`foo` would provide that context like so:

```
type userDefinedKey string
var key userDefinedKey =  "answer"
var answer int64 = 42

foo() {
  ctx := context.WithValue(context.Background(), key, value)
  contextprovider.Provide(ctx, bar)
}
```

and `bar` would consume that context like so:

```
bar() {
  ctx, ok, free := contextprovider.Inject()
}
```

See how there was no need to have a `ctx context.Context` parameter in `bar`? 

If you only want to consume the value without context itself, there's a type-safe way of doing so:

```
bar() {
  val, ok, free := contextprovider.InjectValue[int64](key) // You want a int64 value associated with key
  fmt.Printf("%T %v", val, val) // int64 42
}
```
