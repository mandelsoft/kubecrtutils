package handler

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type MapFunc = handler.TypedMapFunc[client.Object, mcreconcile.Request]

type EventHandler = handler.TypedEventHandler[client.Object, mcreconcile.Request]

func EnqueueRequestFromMapFunc(fn MapFunc) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster cluster.Cluster) EventHandler {
		return handler.TypedEnqueueRequestsFromMapFunc[client.Object, mcreconcile.Request](fn)
	}
}

type MapFuncFactory = TypedMapFuncFactory[client.Object, mcreconcile.Request]

type TypedMapFuncFactory[object client.Object, request comparable] = func(clusterName string, cluster cluster.Cluster) handler.TypedMapFunc[object, request]

func EnqueueRequestFromMapFuncFactory(fn MapFuncFactory) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster cluster.Cluster) EventHandler {
		return handler.TypedEnqueueRequestsFromMapFunc[client.Object, mcreconcile.Request](fn(clusterName, cluster))
	}
}
