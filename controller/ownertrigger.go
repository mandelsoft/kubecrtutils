package controller

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func OwnerTrigger[T client.Object](local ...bool) ResourceTriggerDefinition {
	l := general.OptionalDefaultedBool(true, local...)
	return newTrigger[T](
		ownerMapFuncFactory[client.Object](nil, l),
		"owner trigger",
	)
}

func ownerMapFuncFactory[T client.Object](owmerproto client.Object, local bool) ControllerAware[handler.TypedMapFuncFactory[T, mcreconcile.Request]] {
	return func(ctx context.Context, c types.Controller) (handler.TypedMapFuncFactory[T, mcreconcile.Request], error) {
		if owmerproto == nil {
			owmerproto = c.GetResource()
		}
		gk, err := objutils.GKForObject(c.GetCluster(), owmerproto)
		if err != nil {
			return nil, err
		}
		log := c.GetLogger().WithName("owner").WithName(gk.Kind)
		if local {
			return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[T, mcreconcile.Request] {
				cl := c.GetControllerManager().MapTechnicalName(clusterName)
				return owner.MapOwnerToRequestForGK[T](owner.NewHandler(cl), owner.MatcherFor(c.GetCluster()), cl.AsCluster(), gk, log)
			}, nil
		} else {
			return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[T, mcreconcile.Request] {
				cl := c.GetControllerManager().MapTechnicalName(clusterName)
				return owner.MapOwnerToRequestForGK[T](owner.NewHandler(cl), owner.MatcherFor(c.GetCluster()), cl.AsCluster(), gk, log)
			}, nil
		}
	}
}

func mapOwnersFactory[T client.Object](local bool) ControllerAware[ClusterAware[ObjectMapper[T, owner.Owner]]] {
	return func(ctx context.Context, c types.Controller) (ClusterAware[ObjectMapper[T, owner.Owner]], error) {
		log := c.GetLogger().WithName("owners")
		if local {
			return func(clusterName string, _ cluster.Cluster) ObjectMapper[T, owner.Owner] {
				cl := c.GetControllerManager().MapTechnicalName(clusterName)
				return owner.MapOwners[T](owner.NewHandler(cl), owner.MatcherFor(cl), cl.AsCluster(), log)
			}, nil
		} else {
			return func(clusterName string, _ cluster.Cluster) ObjectMapper[T, owner.Owner] {
				cl := c.GetControllerManager().MapTechnicalName(clusterName)
				return owner.MapOwners[T](owner.NewHandler(cl), owner.MatcherFor(c.GetCluster()), cl.AsCluster(), log)
			}, nil
		}
	}
}
