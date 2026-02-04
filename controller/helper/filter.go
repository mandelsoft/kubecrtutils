package helper

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func FilterCluster[object client.Object](match func(clusterName string) bool, fn handler.TypedMapFunc[object, mcreconcile.Request]) handler.TypedMapFunc[object, reconcile.Request] {
	return func(ctx context.Context, object object) []reconcile.Request {
		return sliceutils.Aggregate(fn(ctx, object), []reconcile.Request{}, func(a []reconcile.Request, r mcreconcile.Request) []reconcile.Request {
			if match != nil && r.ClusterName != "" && !match(r.ClusterName) {
				return a
			}
			return append(a, r.Request)
		})
	}
}
