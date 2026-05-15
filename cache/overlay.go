package cache

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// overlayKey uniquely identifies an object across GVK + namespace/name
type overlayKey struct {
	gvk       schema.GroupVersionKind
	namespace string
	name      string
}

type objectOverlay struct {
	lock    sync.RWMutex
	entries map[overlayKey]client.Object
}

func newObjectOverlay() *objectOverlay {
	return &objectOverlay{
		entries: make(map[overlayKey]client.Object),
	}
}

func (o *objectOverlay) set(obj client.Object) {
	key := overlayKeyFor(obj)
	// Deep copy so mutations to the caller's obj don't corrupt the overlay
	stored := obj.DeepCopyObject().(client.Object)

	o.lock.Lock()
	defer o.lock.Unlock()

	existing, ok := o.entries[key]
	if !ok || isNewer(stored, existing) {
		o.entries[key] = stored
	}
}

// get returns the overlay entry if it exists and matches the type of obj.
func (o *objectOverlay) get(key client.ObjectKey, obj client.Object) (client.Object, bool) {
	okey := overlayKey{
		gvk:       gvkOf(obj),
		namespace: key.Namespace,
		name:      key.Name,
	}

	o.lock.RLock()
	defer o.lock.RUnlock()

	entry, ok := o.entries[okey]
	return entry, ok
}

// evict removes an entry once the cache has caught up to or past its version.
// Call this from Get when the cache version >= overlay version.
func (o *objectOverlay) evict(obj client.Object) {
	key := overlayKeyFor(obj)

	o.lock.Lock()
	defer o.lock.Unlock()

	if existing, ok := o.entries[key]; ok {
		if !isNewer(existing, obj) {
			delete(o.entries, key)
		}
	}
}

func overlayKeyFor(obj client.Object) overlayKey {
	return overlayKey{
		gvk:       gvkOf(obj),
		namespace: obj.GetNamespace(),
		name:      obj.GetName(),
	}
}
