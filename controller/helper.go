package controller

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/controller/helper"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func mapperFactoryForTypedFactory[T client.Object, R comparable](fac handler.TypedControllerAwareMapFuncFactory[T, R], convfactory helper.RequestConverterFactoryForCluster[R]) handler.ControllerAwareMapFuncFactory {
	return func(ctx context.Context, cntr types.Controller) (handler.TypedMapFuncFactory[client.Object, mcreconcile.Request], error) {
		converter := convfactory(ctx, cntr)
		f, err := fac(ctx, cntr)
		if err != nil {
			return nil, err
		}
		return func(clusterName string, cluster sigcluster.Cluster) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
			conv := converter(clusterName)
			m := f(clusterName, cluster)
			return func(ctx context.Context, obj client.Object) []mcreconcile.Request {
				return sliceutils.Transform(m(ctx, (any(obj)).(T)), conv)
			}
		}, nil
	}
}
