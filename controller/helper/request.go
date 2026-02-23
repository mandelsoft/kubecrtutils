package helper

import (
	"context"

	"github.com/mandelsoft/goutils/transformer"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func IdConverter[T any](in T) T {
	return in
}

type RequestConverterForCluster[R comparable] = func(clusterName string) transformer.Transformer[R, mcreconcile.Request]

type RequestConverterFactoryForCluster[R comparable] = func(ct context.Context, controller types.Controller) RequestConverterForCluster[R]

func StaticRequestConverterForCluster[R comparable](converter RequestConverterForCluster[R]) RequestConverterFactoryForCluster[R] {
	return func(ct context.Context, controller types.Controller) RequestConverterForCluster[R] {
		return converter
	}
}

func ConvertMapFunc[O client.Object, R comparable](mapFunc handler.TypedMapFunc[O, R]) handler.TypedMapFunc[client.Object, R] {
	return func(ctx context.Context, object client.Object) []R {
		return mapFunc(ctx, any(object).(O))
	}
}

func LiftRequest(clusterName string) transformer.Transformer[reconcile.Request, mcreconcile.Request] {
	return func(request reconcile.Request) mcreconcile.Request {
		return mcreconcile.Request{
			ClusterName: clusterName,
			Request:     request,
		}
	}
}

// RequestConverterFactoryForClusterCompletion provides a factory for a cluster specific mapper completing an incomplete request
// with the effective cluster name and mapping
// found logical names to effective names used by the underlying queue.
func RequestConverterFactoryForClusterCompletion[R mcreconcile.ClusterAware[R]](cix context.Context, controller types.Controller) func(clusterName string) transformer.Transformer[R, R] {
	mappings := controller.GetClusterMappings()
	// if no mappings are required just skip this mapping step
	if len(mappings) == 0 {
		return CompleteRequest[R]
	}
	return CompleteRequestWithMappings[R](mappings)
}

func CompleteRequest[R mcreconcile.ClusterAware[R]](clusterName string) transformer.Transformer[R, R] {
	return func(request R) R {
		if request.Cluster() != "" {
			return request

		}
		return request.WithCluster(clusterName)
	}
}

func CompleteRequestWithMappings[R mcreconcile.ClusterAware[R]](mappings types.Mappings) func(clusterName string) transformer.Transformer[R, R] {
	return func(clusterName string) transformer.Transformer[R, R] {
		return func(request R) R {
			if request.Cluster() != "" {
				return request.WithCluster(mappings.Map(request.Cluster()))
			}
			return request.WithCluster(clusterName)
		}
	}
}
