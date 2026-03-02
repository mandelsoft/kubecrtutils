package handler

import (
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type ObjectMapper[T any, R any] = types.ObjectMapper[T, R]
type ClusterAware[F any] = types.ClusterAware[F]
type ControllerAware[F any] = types.ControllerAware[F]

type EventHandler = handler.EventHandler
type TypedEventHandler[T any, R comparable] = handler.TypedEventHandler[T, R]

type MapFunc = TypedMapFunc[client.Object, mcreconcile.Request]
type TypedMapFunc[T any, R comparable] = ObjectMapper[T, R] //  handler.TypedMapFunc[T, R] missing = in original definition

type MapFuncFactory = TypedMapFuncFactory[client.Object, mcreconcile.Request]
type TypedMapFuncFactory[object client.Object, request comparable] = ClusterAware[TypedMapFunc[object, request]]

type ControllerAwareMapFuncFactory = TypedControllerAwareMapFuncFactory[client.Object, mcreconcile.Request]
type TypedControllerAwareMapFuncFactory[T client.Object, R comparable] = ControllerAware[TypedMapFuncFactory[T, R]]
type LocalTypedControllerAwareMapFuncFactory[T client.Object] = TypedControllerAwareMapFuncFactory[T, reconcile.Request]

func TypedEnqueueRequestsFromMapFunc[T any, R comparable](fn TypedMapFunc[T, R]) TypedEventHandler[T, R] {
	return handler.TypedEnqueueRequestsFromMapFunc[T, R](fn)
}
