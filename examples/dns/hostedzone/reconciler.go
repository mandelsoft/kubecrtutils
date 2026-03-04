package hostedzone

import (
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
)

type Reconciler struct {
	controller controller.TypedController[*v1alpha1.HostedZone, v1alpha1.HostedZone]
	Runtime    cluster.ClusterEquivalent
	Dataplane  cluster.ClusterEquivalent
}

func (r *Reconciler) Request(def *reconciler.BaseRequest[*v1alpha1.HostedZone]) reconciler.ReconcileRequest[*v1alpha1.HostedZone] {
	return &ReconcileRequest{
		reconciler.DefaultReconcileRequest[*v1alpha1.HostedZone, *Reconciler]{
			BaseRequest: *def,
			Reconciler:  r,
		},
	}
}
