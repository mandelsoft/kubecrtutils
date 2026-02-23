package controller

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/kubedns/pkg/kubecrtutils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func OwnerTrigger[T client.Object](local bool) ResourceTriggerDefinition {
	return newTrigger[T](
		ownerMapFuncFactory[client.Object](nil, local),
		"owner trigger",
	)
}

func ownerMapFuncFactory[T client.Object](owmerproto client.Object, local bool) ControllerAware[TypedMapFuncFactory[T, mcreconcile.Request]] {
	return func(ctx context.Context, c types.Controller) (TypedMapFuncFactory[T, mcreconcile.Request], error) {
		if owmerproto == nil {
			owmerproto = c.GetResource()
		}
		gk, err := kubecrtutils.GKForObject(c.GetCluster(), owmerproto)
		if err != nil {
			return nil, err
		}
		log := c.GetLogger().WithName("owner").WithName(gk.Kind)
		if local {
			return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[T, mcreconcile.Request] {
				cl := c.GetControllerManager().MapTechnicalName(clusterName)
				return owner.MapOwnerToRequestForGK[T](owner.NewHandler(cl), owner.MatcherFor(cl), cl.AsCluster(), gk, log)
			}, nil
		} else {
			return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[T, mcreconcile.Request] {
				cl := c.GetControllerManager().MapTechnicalName(clusterName)
				return owner.MapOwnerToRequestForGK[T](owner.NewHandler(cl), owner.MatcherFor(c.GetCluster()), cl.AsCluster(), gk, log)
			}, nil
		}
	}
}

func OwnerMapFuncFactory[T, O client.Object](local bool) ControllerAware[TypedMapFuncFactory[T, mcreconcile.Request]] {
	return ownerMapFuncFactory[T](generics.ObjectFor[O](), local)
}
