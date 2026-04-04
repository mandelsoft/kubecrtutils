package cacheindex

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TypedIndices[T any] interface {
	internal.Group[TypedIndex[T]]
}

func NewTypedIndices[T any](name string) TypedIndices[T] {
	return internal.NewGroup[TypedIndex[T]](name)
}

func GetTypedIndex[T any](indices Indices, name string) TypedIndex[T] {
	i := indices.Get(name)
	if i == nil {
		return nil
	}
	return generics.Cast[TypedIndex[T]](i.GetEffective())
}

type TypedIndex[T any] interface {
	types.Index

	GetTyped(ctx context.Context, ns string, key string) ([]T, error)
	ForEachTypedItem(ctx context.Context, namespace, key string, action func(object *T) error) error
}

type _typedIndex[T any] struct {
	types.Index
}

var _ TypedIndex[any] = (*_typedIndex[any])(nil)

func (i *_typedIndex[T]) GetEffective() Index {
	return i
}

func (i *_typedIndex[T]) GetTyped(ctx context.Context, namespace string, key string) ([]T, error) {
	list, err := i.GetList(ctx, namespace, key)
	if err != nil {
		return nil, err
	}
	return GetItemList[T](list)
}

func (i *_typedIndex[T]) ForEachTypedItem(ctx context.Context, namespace, key string, action func(object *T) error) error {
	list, err := i.GetTyped(ctx, namespace, key)
	if err != nil {
		return err
	}
	for _, e := range list {
		err := action(&e)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetItemList[T any](list client.ObjectList) ([]T, error) {
	ptr, err := meta.GetItemsPtr(list)
	if err != nil {
		return nil, err
	}
	return *(ptr).(*[]T), nil
}

type IndexProvider interface {
	GetIndex(name string) Index
}

func GetIndexFrom[T any](provider IndexProvider, name string) (TypedIndex[T], error) {
	i := provider.GetIndex(name)
	if i == nil {
		return nil, fmt.Errorf("no index for %q found", name)
	}
	t, ok := i.(TypedIndex[T])
	if !ok {
		var e T
		return nil, fmt.Errorf("type mismatch for %q: %T expected, but found %T", name, &e, i.GetResource())
	}
	return t, nil
}
