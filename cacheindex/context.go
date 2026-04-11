package cacheindex

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/generics"
)

func OptionsFromContext(ctx context.Context) flagutils.Options {
	o := generics.Cast[flagutils.Options](ctx.Value("options"))
	if o == nil {
		return nil
	}
	return flagutils.Unwrap(o)
}
