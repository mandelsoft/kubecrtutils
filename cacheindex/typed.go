package cacheindex

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TypedIndex[T any] interface {
	types.Index

	GetTyped(ctx context.Context, ns string, key string) ([]T, error)
	ForEachTypedItem(ctx context.Context, namespace, key string, action func(object *T) error) error
}

type _typedIndex[T any] struct {
	types.Index
}

var _ TypedIndex[any] = (*_typedIndex[any])(nil)

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
