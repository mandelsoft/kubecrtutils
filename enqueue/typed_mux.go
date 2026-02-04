package enqueue

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type TypedMux[T comparable] interface {
	TriggerSource(obj runtime.Object) (TypedEnqueue[T], error)
	Mux
}

type RequestFunc[T comparable] func(ctx context.Context, req client.ObjectKey) (T, error)

type typedmux[T comparable] struct {
	lock     sync.Mutex
	scheme   *runtime.Scheme
	enqueues map[schema.GroupVersionKind]TypedEnqueue[T]
	creator  RequestFunc[T]
}

func NewTypedMux[T comparable](scheme *runtime.Scheme, creator RequestFunc[T]) TypedMux[T] {
	return &typedmux[T]{scheme: scheme, creator: creator, enqueues: make(map[schema.GroupVersionKind]TypedEnqueue[T])}
}

func (m *typedmux[T]) TriggerSource(obj runtime.Object) (TypedEnqueue[T], error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	gvk, err := apiutil.GVKForObject(obj, m.scheme)
	if err != nil {
		return nil, err
	}
	e := m.enqueues[gvk]
	if e == nil {
		e = NewTypedEnqueue[T]()
		m.enqueues[gvk] = e
	}
	return e, nil
}

func (m *typedmux[T]) EnqueueByGVK(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	e := m.enqueues[gvk]
	if e != nil {
		req, err := m.creator(ctx, key)
		if err != nil {
			return err
		}
		e.AddToQueue(req)
	}
	return nil
}

func (m *typedmux[T]) EnqueueByObject(ctx context.Context, obj runtime.Object) error {
	var err error
	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Empty() {
		gvk, err = apiutil.GVKForObject(obj, m.scheme)
		if err != nil {
			return err
		}
	}
	k, err := GetKey(obj)
	if err != nil {
		return err
	}
	m.EnqueueByGVK(ctx, gvk, k)
	return nil
}
