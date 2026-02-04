package controller

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func typedEnqueueRequestsFromMapFunc[object client.Object, request mcreconcile.ClusterAware[request]](fn handler.TypedMapFunc[object, request]) mchandler.TypedEventHandlerFunc[object, request] {
	return func(clusterName string, cl sigcluster.Cluster) handler.TypedEventHandler[object, request] {
		return handler.TypedEnqueueRequestsFromMapFunc[object, request](withClusterCompletion(clusterName, fn))
	}
}

func withClusterCompletion[object client.Object, request mcreconcile.ClusterAware[request]](clusterName string, fn handler.TypedMapFunc[object, request]) handler.TypedMapFunc[object, request] {
	return func(ctx context.Context, object object) []request {
		return sliceutils.Transform(fn(ctx, object), func(r request) request {
			if r.Cluster() != "" {
				return r
			}
			return r.WithCluster(clusterName)
		})
	}
}

func withClusterInjection[object client.Object](clusterName string, fn handler.TypedMapFunc[object, reconcile.Request]) handler.TypedMapFunc[object, mcreconcile.Request] {
	return func(ctx context.Context, object object) []mcreconcile.Request {
		return sliceutils.Transform(fn(ctx, object), func(r reconcile.Request) mcreconcile.Request {
			return mcreconcile.Request{
				Request:     r,
				ClusterName: clusterName,
			}
		})
	}
}
