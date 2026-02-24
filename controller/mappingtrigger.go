package controller

import (
	"github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/controller/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func ResourceTrigger[T client.Object](mapFunc handler.TypedMapFunc[T, mcreconcile.Request], desc ...string) ResourceTriggerDefinition {
	return ResourceTriggerByFactory[T](handler.Lift(mapFunc), desc...)
}

func ResourceTriggerByFactory[T client.Object](mapFunc handler.TypedControllerAwareMapFuncFactory[T, mcreconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTrigger[T](
		mapperFactoryForTypedFactory[T, mcreconcile.Request](
			mapFunc,
			helper.RequestConverterFactoryForClusterCompletion,
		),
		desc...,
	)
}

/////////////////////////////////////////////////ResourceTriggerDefinition///////////////////////////////

func LocalResourceTrigger[T client.Object](mapFunc handler.TypedMapFunc[T, reconcile.Request], desc ...string) ResourceTriggerDefinition {
	return LocalResourceTriggerByFactory[T](
		handler.Lift(mapFunc),
		desc...,
	)
}

func LocalResourceTriggerByFactory[T client.Object](mapper handler.LocalTypedControllerAwareMapFuncFactory[T], desc ...string) ResourceTriggerDefinition {
	return newTrigger[T](
		mapperFactoryForTypedFactory[T, reconcile.Request](mapper, helper.StaticRequestConverterForCluster(helper.LiftRequest)),
		desc...,
	)
}
