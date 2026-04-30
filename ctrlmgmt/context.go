package ctrlmgmt

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
)

func FromContext(ctx context.Context) ControllerManager {
	return generics.Cast[ControllerManager](ctx.Value("controller-manager"))
}

func addToContext(ctx context.Context, c ControllerManager) context.Context {
	return context.WithValue(ctx, "controller-manager", c)
}
