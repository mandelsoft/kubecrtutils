package cache

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WriteThroughCache wraps a controller-runtime client.
// Writes go through to the API server as normal.
// The result (mutated obj) is stored locally in a version-aware overlay.
// Get checks the overlay first — if the overlay has a newer ResourceVersion
// than the cache, it returns the overlay copy; otherwise falls through to cache.
type WriteThroughCache struct {
	client.Client
	overlay *objectOverlay
}

func NewWriteThroughClient(c client.Client) *WriteThroughCache {
	return &WriteThroughCache{
		Client:  c,
		overlay: newObjectOverlay(),
	}
}

// Get serves from the overlay if it holds a newer version, otherwise
// falls through to the underlying cached client.
func (w *WriteThroughCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	cacheErr := w.Client.Get(ctx, key, obj, opts...)

	overlayObj, hasOverlay := w.overlay.get(key, obj)

	if !hasOverlay {
		// Nothing in overlay — just return whatever cache gave us
		return cacheErr
	}

	if cacheErr != nil {
		// Cache doesn't have it yet — serve overlay
		copyInto(overlayObj, obj)
		return nil
	}

	// Both exist — compare versions
	if isNewer(overlayObj, obj) {
		// Overlay is ahead of cache — serve overlay
		copyInto(overlayObj, obj)
	} else {
		// Cache has caught up or overtaken overlay — evict
		w.overlay.evict(obj)
		// obj already holds the cache result, nothing more to do
	}

	return nil
}

func (w *WriteThroughCache) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := w.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	w.overlay.set(obj)
	return nil
}

func (w *WriteThroughCache) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := w.Client.Update(ctx, obj, opts...); err != nil {
		return err
	}
	w.overlay.set(obj)
	return nil
}

func (w *WriteThroughCache) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if err := w.Client.Patch(ctx, obj, patch, opts...); err != nil {
		return err
	}
	w.overlay.set(obj)
	return nil
}
