package objutils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AnnotationModifier = func(map[string]string) map[string]string

func SetAnnotation(obj metav1.Object, key, value string) {
	values := obj.GetAnnotations()
	if values == nil {
		values = map[string]string{}
	}
	values[key] = value
	obj.SetAnnotations(values)
}

func ModifyAnnotations(obj metav1.Object, mod AnnotationModifier) {
	if mod != nil {
		obj.SetAnnotations(mod(obj.GetAnnotations()))
	}
}

func CheckAnnotation(obj metav1.Object, key string, v string) bool {
	values := obj.GetAnnotations()
	if values == nil {
		return false
	}

	e, ok := values[key]
	return ok && e == v
}

func GetAnnotation(obj metav1.Object, key string) string {
	values := obj.GetAnnotations()
	if values == nil {
		return ""
	}
	return values[key]
}

func RemoveAnnotation(obj metav1.Object, key string) bool {
	values := obj.GetAnnotations()
	if values == nil {
		return false
	}
	_, ok := values[key]
	delete(values, key)
	return ok
}
