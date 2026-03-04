package objutils

import (
	"iter"

	"github.com/mandelsoft/goutils/generics"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateObjectList(obj runtime.Object, scheme *runtime.Scheme) (client.ObjectList, error) {
	listGVK, err := GetListGVK(obj, scheme)
	if err != nil {
		return nil, err
	}
	o, err := scheme.New(listGVK)
	if err != nil {
		return nil, err
	}
	return o.(client.ObjectList), nil
}

func ObjectListFor[T client.Object](scheme *runtime.Scheme) (client.ObjectList, error) {
	return CreateObjectList(generics.ObjectFor[T](), scheme)
}

func ObjectListLen(obj client.ObjectList) int {
	itemsPtr, err := meta.GetItemsPtr(obj)
	if err != nil {
		return 0
	}
	items, err := conversion.EnforcePtr(itemsPtr)
	if err != nil {
		return 0
	}
	if items.IsNil() {
		return 0
	}
	return items.Len()
}

func Items(list client.ObjectList) iter.Seq2[int, client.Object] {
	return func(yield func(int, client.Object) bool) {
		items, err := meta.ExtractList(list)
		if err != nil {
			return
		}
		for i, item := range items {
			if !yield(i, item.(client.Object)) {
				return
			}
		}
	}
}
