package cache

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// isNewer returns true if a has a strictly greater ResourceVersion than b.
func isNewer(a, b client.Object) bool {
	return a.GetResourceVersion() > b.GetResourceVersion()
}

// copyInto copies the content of src into dst, preserving dst's type.
func copyInto(src, dst client.Object) {
	srcCopy := src.DeepCopyObject().(client.Object)
	// Reflect the internal fields across — the cleanest way is via
	// the unstructured conversion path.
	dst.SetUID(srcCopy.GetUID())
	dst.SetResourceVersion(srcCopy.GetResourceVersion())
	dst.SetGeneration(srcCopy.GetGeneration())
	dst.SetAnnotations(srcCopy.GetAnnotations())
	dst.SetLabels(srcCopy.GetLabels())
	dst.SetFinalizers(srcCopy.GetFinalizers())
	// For full spec/status copy you need scheme-aware conversion — see note below
}

// gvkOf extracts the GVK from an object's TypeMeta if set.
// In practice you'd use apiutil.GVKForObject(obj, scheme) for reliability.
func gvkOf(obj client.Object) schema.GroupVersionKind {
	gvks := obj.GetObjectKind().GroupVersionKind()
	return gvks
}
