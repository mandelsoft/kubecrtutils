package owner

import (
	"context"

	"github.com/mandelsoft/kubecrtutils"
	clusterutils "github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func MapOwnerToRequestByObject(hdlr Handler, matcher ClusterMatcher, p clusterutils.SchemeProvider, src types.Cluster, obj client.Object, log ...logging.Logger) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
	gk, err := kubecrtutils.GKForObject(p, obj)
	if err != nil {
		panic(err)
	}
	return MapOwnerToRequest(hdlr, matcher, src, gk, log...)
}

func MapOwnerToRequest(hdlr Handler, matcher ClusterMatcher, src types.Cluster, kind schema.GroupKind, log ...logging.Logger) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
	return func(ctx context.Context, obj client.Object) []mcreconcile.Request {
		cl := src
		if cl == nil {
			cl = clustercontext.ClusterFor(ctx)
		}
		cname, okey := hdlr.GetOwner(matcher, cl, obj, kind)
		if okey == nil {
			return nil
		}
		if len(log) > 0 {
			log[0].Info("trigger owner {{owner}} of modified object {{modified}}",
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

/*
func WatchSourceForSlave[O, R client.Object](c clusterutils.Cluster, owner OwnerHandler, s clusterutils.SchemeProvider, log ...logging.Logger) source.Source {
	// O,R are pointer types, but we need an object

	o := reflect.New(generics.TypeOf[O]().Elem()).Interface().(O)
	r := reflect.New(generics.TypeOf[R]().Elem()).Interface().(R)

	return source.Kind(c.GetCache(), o,
		handler.TypedEnqueueRequestsFromMapFunc[O, reconcile.Request](MapOwnerToLocalRequestByObject[O](owner, s, r, log...)))

}

*/
