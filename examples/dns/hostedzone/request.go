package hostedzone

import (
	"github.com/kcp-dev/logicalcluster/v3"
	"github.com/mandelsoft/goutils/funcs"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/kcp"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcileRequest struct {
	reconciler.DefaultReconcileRequest[*v1alpha1.HostedZone, *Reconciler]
}

func (r *ReconcileRequest) Reconcile() reconcile.Problem {
	// r.Info("handling local namespace {{namespace}}", "namespace", r.GenerateNameFor("runtime", "demo"))
	other := &v1alpha1.HostedZone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other",
			Namespace: r.Namespace,
		},
	}
	r.Info("URL {{url}}", "url", funcs.First(r.GetAPIServerURL()))

	var list v1alpha1.HostedZoneList
	p := r.Reconciler.controller.GetCluster().AsFleet().(*kcp.Fleet).GetKCPProvider()

	for n, pp := range p.Providers {
		r.Info("lookup shard {{shard}}", "shard", n)
		c := pp.GetCache()
		err := c.List(r.Context, &list, client.MatchingFields{"IndexKeyZoneParent": "*/" + r.Name})
		// err := r.List(r.Context, &list, client.MatchingFields{"IndexKeyZoneParent": r.Name})
		if err == nil {
			for _, e := range list.Items {
				r.Info("  found nested: {{cluster}} {{nested}}", "cluster", logicalcluster.From(&e), "nested", e.Name)
			}
		}
	}

	var secret v1.Secret

	err := r.Cluster.GetCluster().GetClient().Get(r.Context, client.ObjectKey{Namespace: r.Namespace, Name: "dns-service"}, &secret)
	if err == nil {
		r.Info("ca.crt", "cert", secret.Data["ca.crt"])
	} else {
		r.Error("cannot get secret", "error", err)
	}
	_ = other
	r.EnqueueByObject(r.Context, other)
	return reconcile.Succeeded()
}
