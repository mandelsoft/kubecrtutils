package hostedzone

import (
	"github.com/mandelsoft/kubecrtutils/controller"
	corednsv1alpha1 "github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
)

func Controller() controller.Definition {
	return controller.Define[*corednsv1alpha1.HostedZone]("coredns.mandelsoft.hostedzone", "dataplane", &ReconcilerFactory{}).
		// UseCluster("runtime").
		AddIndex("IndexKeyZoneParent", parentIndexer)
	// AddTrigger(
	//   controller.OwnerTrigger[*appsv1.Deployment]().OnCluster("runtime"),
	//   controller.OwnerTrigger[*corev1.Secret]().OnCluster("runtime"),
	// )
}

func parentIndexer(o *corednsv1alpha1.HostedZone) []string {
	if o.Spec.ParentRef == "" {
		return nil
	}
	return []string{o.Spec.ParentRef}
}
