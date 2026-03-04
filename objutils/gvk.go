package objutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func GetListGVK(obj runtime.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	// 1. Determine GVK
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	// 2. Some schemes might have specific logic to return the list type
	// but usually, we just use the naming convention:
	listGVK := gvk
	listGVK.Kind = gvk.Kind + "List"

	// 3. Verify the List GVK actually exists in the scheme
	if !scheme.Recognizes(listGVK) {
		return schema.GroupVersionKind{}, fmt.Errorf("list kind %s not recognized", listGVK.Kind)
	}

	return listGVK, nil
}
