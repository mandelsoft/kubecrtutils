package controller

import (
	"context"

	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func OwnerTrigger[T client.Object]() ResourceTriggerDefinition {
	return newTrigger[T](
		ownerFunc[T],
		"owner trigger",
	)
}

func ownerFunc[T client.Object](ctx context.Context, c types.Controller) (myhandler.MapFuncFactory, error) {
	return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
		cl := c.GetControllerManager().MapTechnicalName(clusterName)
		return owner.MapOwnerToRequestByObject(owner.NewHandler(cl), owner.MatcherFor(c.GetCluster()), c.GetCluster(), cl.AsCluster(), c.GetResource())
	}, nil
}
