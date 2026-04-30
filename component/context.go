package component

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
)

func FromContext(ctx context.Context) Definition {
	return generics.Cast[Definition](ctx.Value("component"))
}

func addToContext(ctx context.Context, c Definition) context.Context {
	return context.WithValue(ctx, "component", c)
}
