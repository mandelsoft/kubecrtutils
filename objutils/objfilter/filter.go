package objfilter

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Interface interface {
	Filter(obj client.Object) bool
}

type Func func(obj client.Object) bool

func (f Func) Filter(obj client.Object) bool {
	return f(obj)
}

func MetaGroupKind(gk metav1.GroupKind) Interface {
	return GroupKind(gk.Group, gk.Kind)
}

func SchemaGroupKind(gk schema.GroupKind) Interface {
	return GroupKind(gk.Group, gk.Kind)
}

func GroupKind(group, kind string) Interface {
	return Func(func(obj client.Object) bool {
		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk.Group == "core" {
			gvk.Group = ""
		}
		if group == "core" {
			group = ""
		}
		return gvk.Group == group && gvk.Kind == kind
	})
}

func Or(filters ...Interface) Interface {
	return Func(func(obj client.Object) bool {
		for _, f := range filters {
			if f.Filter(obj) {
				return true
			}
		}
		return false
	})
}

func And(filters ...Interface) Interface {
	return Func(func(obj client.Object) bool {
		for _, f := range filters {
			if !f.Filter(obj) {
				return false
			}
		}
		return true
	})
}

func Not(filter Interface) Interface {
	return Func(func(obj client.Object) bool {
		return !filter.Filter(obj)
	})
}
