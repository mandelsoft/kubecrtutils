package ctxutils

import (
	"context"
)

func WithCancel[T comparable](key T, ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	return context.WithValue(ctx, key, cancel)
}

func Cancel[T comparable](key T, ctx context.Context) bool {
	cancel := ctx.Value(key)
	if cancel == nil {
		return false
	}
	cancel.(context.CancelFunc)()
	return true
}
