package hostedzone

import (
	"github.com/kcp-dev/logicalcluster/v3"
	"github.com/mandelsoft/goutils/funcs"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/kcp"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcileRequest struct {
	reconciler.DefaultReconcileRequest[*v1alpha1.HostedZone, *Reconciler]
}

func (r *ReconcileRequest) Reconcile() reconcile.Problem {
	o := r.GetController().GetOptions()
	_ = o
	ns := objutils.GenerateUniqueName("dns-service", r.Cluster.GetId(), "", r.Namespace, objutils.MAX_NAMESPACELEN)
	r.Info("handling local namespace {{namespace}}", "namespace", ns)
	other := &v1alpha1.HostedZone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other",
			Namespace: r.Namespace,
		},
	}
	r.Info("URL {{url}}", "url", funcs.First(r.GetAPIServerURL()))

	r.Info("global index for {{name}}", "name", r.Name)
	list, err := r.GetController().GetCluster().ListIndexedGlobalKeys(r.Context, r.GetController().GetResource(), "IndexKeyZoneParent", r.Name)
	if err == nil {
		for _, e := range list {
			r.Info("  found nested: {{cluster}} {{nested}}", "cluster", e.ClusterName, "nested", e.NamespacedName)
		}
	} else {
		r.Error("index error: {{error}}", "error", err)
	}

	{
		r.Info("technical index implementation for {{name}}", "name", r.Name)
		var list v1alpha1.HostedZoneList
		p := r.Reconciler.controller.GetCluster().AsFleet().(*kcp.Fleet).GetKCPProvider()
		for n, pp := range p.Providers {
			r.Info("lookup shard {{shard}}", "shard", n)
			c := pp.GetCache()
			err := c.List(r.Context, &list, client.MatchingFields{"IndexKeyZoneParent": "*/" + r.Name})
			// err := r.List(r.Context, &list, client.MatchingFields{"IndexKeyZoneParent": r.Name})
			if err == nil {
				for _, e := range list.Items {
					r.Info("  found nested: {{cluster}} {{nested}}", "cluster", logicalcluster.From(&e), "nested", e.GetName())
				}
			} else {
				r.Error("index error: {{error}}", "error", err)
			}
		}
	}

	var secret v1.Secret

	err = r.Cluster.Get(r.Context, client.ObjectKey{Namespace: r.Namespace, Name: "dns-service"}, &secret)
	if err == nil {
		r.Info("ca.crt", "cert", secret.Data["ca.crt"])
	} else {
		r.Error("cannot get secret", "error", err)
	}

	err = r.GetController().GetLogicalCluster("runtime").Get(r.Context, client.ObjectKey{Namespace: ns, Name: "dns-service"}, &secret)
	if err == nil {
		r.Info("kubeconfig", "cert", string(secret.Data["KubeConfig"]))
	} else {
		r.Error("cannot get secret", "error", err, "namespace", ns, "name", "dns-service")
	}
	_ = other
	r.EnqueueByObject(r.Context, other)
	return reconcile.Succeeded()
}
