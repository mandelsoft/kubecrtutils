package helper

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type Converter[I, O any] = func(I) O

type RequestConverterForCluster[R comparable] = func(clusterName string) Converter[R, mcreconcile.Request]

func ConvertMapFunc[O client.Object, R comparable](mapFunc handler.TypedMapFunc[O, R]) handler.TypedMapFunc[client.Object, R] {
	return func(ctx context.Context, object client.Object) []R {
		return mapFunc(ctx, any(object).(O))
	}
}

func LiftRequest(clusterName string) Converter[reconcile.Request, mcreconcile.Request] {
	return func(request reconcile.Request) mcreconcile.Request {
		return mcreconcile.Request{
			ClusterName: clusterName,
			Request:     request,
		}
	}
}

func CompleteRequest[R mcreconcile.ClusterAware[R]](clusterName string) Converter[R, R] {
	return func(request R) R {
		if request.Cluster() != "" {
			return request
		}
		return request.WithCluster(clusterName)
	}
}
