package hostedzone

import (
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReconcileRequest struct {
	reconciler.DefaultReconcileRequest[*v1alpha1.HostedZone, *Reconciler]
}

func (r *ReconcileRequest) Reconcile() reconcile.Problem {
	r.Info("handling local namespace {{namespace}}", "namespace", r.GenerateNameFor("runtime", "demo"))
	other := &v1alpha1.HostedZone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other",
			Namespace: r.Namespace,
		},
	}
	_ = other
	r.EnqueueByObject(r.Context, other)
	return reconcile.Succeeded()
}
