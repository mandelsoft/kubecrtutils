package enqueue

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Mux interface {
	EnqueueByGVK(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) error
	EnqueueByObject(ctx context.Context, obj runtime.Object) error
}

func NewMux(scheme *runtime.Scheme) TypedMux[reconcile.Request] {
	return NewTypedMux[reconcile.Request](scheme, func(_ context.Context, key client.ObjectKey) (reconcile.Request, error) {
		return reconcile.Request{
			NamespacedName: key,
		}, nil
	})
}

func GetKey(obj runtime.Object) (client.ObjectKey, error) {
	// runtime.Object is an interface that doesn't strictly guarantee
	// access to metadata. client.Object adds those methods.
	accessor, ok := obj.(client.Object)
	if !ok {
		return client.ObjectKey{}, fmt.Errorf("object does not implement client.Object")
	}

	return client.ObjectKey{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
	}, nil
}
