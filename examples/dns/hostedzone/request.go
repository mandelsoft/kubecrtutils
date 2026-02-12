package hostedzone

import (
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
)

type ReconcileRequest struct {
	reconciler.DefaultReconcileRequest[*v1alpha1.HostedZone, *Reconciler]
}

func (r *ReconcileRequest) Reconcile() reconcile.Problem {
	r.Info("handling local namespace {{namespace}}", "namespace", r.GenerateNameFor("runtime", "demo"))
	return reconcile.Succeeded()
}
