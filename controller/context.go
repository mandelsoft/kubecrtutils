package controller

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
)

func ControllerFromContext(ctx context.Context) Controller {
	return generics.Cast[Controller](ctx.Value("controller"))
}
