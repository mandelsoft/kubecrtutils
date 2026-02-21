package handler

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/controller/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func MapFuncFactoryWithClusterCompletion(fn handler.TypedMapFunc[client.Object, mcreconcile.Request]) MapFuncFactory {
	return TypedMapFuncFactoryWithClusterCompletion[client.Object, mcreconcile.Request](fn)
}

func TypedMapFuncFactoryWithClusterCompletion[object client.Object, request mcreconcile.ClusterAware[request]](fn handler.TypedMapFunc[object, request]) TypedMapFuncFactory[object, request] {
	return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[object, request] {
		return func(ctx context.Context, obj object) []request {
			return sliceutils.Transform(fn(ctx, obj), func(r request) request {
				if r.Cluster() != "" {
					return r
				}
				return r.WithCluster(clusterName)
			})
		}
	}
}

func MapFuncFactoryWithClusterInjection(fn handler.MapFunc) MapFuncFactory {
	return TypedMapFuncFactoryWithClusterInjection[client.Object](fn)

}
func TypedMapFuncFactoryWithClusterInjection[object client.Object](fn handler.TypedMapFunc[object, reconcile.Request]) TypedMapFuncFactory[object, mcreconcile.Request] {
	return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[object, mcreconcile.Request] {
		c := helper.LiftRequest(clusterName)
		return func(ctx context.Context, obj object) []mcreconcile.Request {
			return sliceutils.Transform(fn(ctx, obj), c)
		}
	}
}
