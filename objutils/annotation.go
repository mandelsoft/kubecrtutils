package objutils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AnnotationModifier func(map[string]string) map[string]string

func SetAnnotation(obj metav1.Object, key, value string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[key] = value
	obj.SetAnnotations(annotations)
}

func ModifyAnnotations(obj metav1.Object, mod AnnotationModifier) {
	if mod != nil {
		obj.SetAnnotations(mod(obj.GetAnnotations()))
	}
}

func GetAnnotation(obj metav1.Object, key string) string {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return ""
	}
	return annotations[key]
}
