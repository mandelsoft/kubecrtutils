package owner

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func MapOwnerToRequestForGK[T client.Object](hdlr Handler, matcher ClusterMatcher, src types.Cluster, kind schema.GroupKind, log ...logging.Logger) handler.TypedMapFunc[T, mcreconcile.Request] {
	return func(ctx context.Context, obj T) []mcreconcile.Request {
		cl := src
		if cl == nil {
			cl = clustercontext.ClusterFor(ctx)
		}
		cname, okey := hdlr.GetOwner(matcher, cl, obj, kind)
		if okey == nil {
			return nil
		}
		if len(log) > 0 {
			log[0].Info("found owner {{owner}} of modified object {{modified}}",
				"owner", *okey,
				"modified", client.ObjectKeyFromObject(obj))
		}
		return []mcreconcile.Request{
			{ClusterName: cname,
				Request: reconcile.Request{
					NamespacedName: *okey,
				},
			},
		}
	}
}

func MapLocalOwnerToRequestForGK(hdlr Handler, matcher ClusterMatcher, src types.Cluster, kind schema.GroupKind, log ...logging.Logger) handler.TypedMapFunc[client.Object, reconcile.Request] {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		cl := src
		if cl == nil {
			cl = clustercontext.ClusterFor(ctx)
		}
		cname, okey := hdlr.GetOwner(matcher, cl, obj, kind)
		if okey == nil || cname != cl.GetName() {
			return nil
		}
		if len(log) > 0 {
			log[0].Info("owner of object {{modified}} in {{cluster}} is {{owner}}",
				"owner", *okey,
				"modified", client.ObjectKeyFromObject(obj))
		}
		return []reconcile.Request{
			{NamespacedName: *okey},
		}
	}
}
