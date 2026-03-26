package objutils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabelModifier = func(map[string]string) map[string]string

func Setlabel(obj metav1.Object, key, value string) {
	values := obj.GetAnnotations()
	if values == nil {
		values = map[string]string{}
	}
	values[key] = value
	obj.SetAnnotations(values)
}

func ModifyLabels(obj metav1.Object, mod LabelModifier) {
	if mod != nil {
		obj.SetLabels(mod(obj.GetLabels()))
	}
}

func GetLabel(obj metav1.Object, key string) string {
	values := obj.GetLabels()
	if values == nil {
		return ""
	}
	return values[key]
}

func RemoveLabel(obj metav1.Object, key string) bool {
	values := obj.GetLabels()
	if values == nil {
		return false
	}
	_, ok := values[key]
	delete(values, key)
	return ok
}
