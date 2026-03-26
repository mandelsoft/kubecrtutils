package objutils

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CleanupMeta(obj client.Object) {
	obj.SetUID("")
	obj.SetResourceVersion("")
	RemoveAnnotation(obj, "managedFields")
	obj.SetDeletionTimestamp(nil)
	CleanupKCPInfo(obj)
}

func CleanupKCPInfo(obj client.Object) {
	if annos := obj.GetAnnotations(); len(annos) > 0 {
		for k, v := range annos {
			if strings.HasSuffix(v, ".kcp.io") {
				delete(annos, k)
			}
		}
	}
	if labels := obj.GetLabels(); len(labels) > 0 {
		for k, v := range labels {
			if strings.HasSuffix(v, ".kcp.io") {
				delete(labels, k)
			}
		}
	}
}
