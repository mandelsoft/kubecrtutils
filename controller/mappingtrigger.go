package controller

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func ResourceTrigger[T client.Object](mapFunc handler.TypedMapFunc[T, mcreconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTriggerF[T](
		myhandler.MapFuncFactoryWithClusterCompletion(ConvertMapFunc[T, mcreconcile.Request](mapFunc)),
		desc...)
}

func ResourceTriggerByFactory[T client.Object](mapper TypedMapperFactory[T, mcreconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTrigger[T](mapperFactoryForTypedFactory[T, mcreconcile.Request](mapper, CompleteRequest), desc...)
}

/////////////////////////////////////////////////ResourceTriggerDefinition///////////////////////////////

func LocalResourceTrigger[T client.Object](mapFunc handler.TypedMapFunc[T, reconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTriggerF[T](
		myhandler.MapFuncFactoryWithClusterInjection(ConvertMapFunc[T, reconcile.Request](mapFunc)),
		desc...)
}

func LocalResourceTriggerByFactory[T client.Object](mapper LocalTypedMapperFactory[T], desc ...string) ResourceTriggerDefinition {
	return newTrigger[T](mapperFactoryForTypedFactory(mapper, LiftRequest), desc...)
}

func withClusterCompletion[object client.Object, request mcreconcile.ClusterAware[request]](clusterName string, fn handler.TypedMapFunc[object, request]) handler.TypedMapFunc[object, request] {
	f := CompleteRequest[request](clusterName)
	return func(ctx context.Context, object object) []request {
		return sliceutils.Transform(fn(ctx, object), f)
	}
}

func withClusterInjection[object client.Object](clusterName string, fn handler.TypedMapFunc[object, reconcile.Request]) handler.TypedMapFunc[object, mcreconcile.Request] {
	f := LiftRequest(clusterName)
	return func(ctx context.Context, object object) []mcreconcile.Request {
		return sliceutils.Transform(fn(ctx, object), f)
	}
}
