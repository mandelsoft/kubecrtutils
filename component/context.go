package component

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
)

func ComponentFromContext(ctx context.Context) Component {
	return generics.Cast[Component](ctx.Value("component"))
}
