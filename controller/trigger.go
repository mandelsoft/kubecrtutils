package controller

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type MapFunc = TypedMapFunc[client.Object, mcreconcile.Request]
type TypedMapFunc[T client.Object, R comparable] = handler.TypedMapFunc[T, R]

type MapFuncFactory = TypedMapFuncFactory[client.Object, mcreconcile.Request]
type TypedMapFuncFactory[object client.Object, request comparable] = ClusterAware[handler.TypedMapFunc[object, request]]

type ControllerAwareMapFuncFactory = TypedControllerAwareMapFuncFactory[client.Object, mcreconcile.Request]
type TypedControllerAwareMapFuncFactory[T client.Object, R comparable] = ControllerAware[TypedMapFuncFactory[T, R]]
type LocalTypedControllerAwareMapFuncFactory[T client.Object] = TypedControllerAwareMapFuncFactory[T, reconcile.Request]

func Lift[F any](f F) ControllerAware[ClusterAware[F]] {
	return LiftToController(LiftToCluster(f))
}

func LiftToCluster[F any](f F) ClusterAware[F] {
	return func(clusterName string, cluster sigcluster.Cluster) F {
		return f
	}
}

func LiftToController[F any](f F) ControllerAware[F] {
	return func(ctx context.Context, c Controller) (F, error) {
		return f, nil
	}
}

type ResourceTriggerDefinition interface {
	OnCluster(name string) ResourceTriggerDefinition

	GetDescription() string
	GetResource() client.Object
	GetMapper() ControllerAwareMapFuncFactory
	GetCluster() string
	Error() error
}

type _trigger struct {
	desc    string
	proto   client.Object
	mapper  ControllerAwareMapFuncFactory
	cluster string
	err     error
}

func newTrigger[T client.Object](mapper ControllerAwareMapFuncFactory, desc ...string) *_trigger {
	return &_trigger{
		desc:   general.OptionalDefaulted("resource mapping", desc...),
		proto:  generics.ObjectFor[T](),
		mapper: mapper,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (t *_trigger) OnCluster(cluster string) ResourceTriggerDefinition {
	t.cluster = cluster
	return t
}

func (t *_trigger) GetResource() client.Object {
	return t.proto
}

func (t *_trigger) GetDescription() string {
	return t.desc
}

func (t *_trigger) GetMapper() ControllerAwareMapFuncFactory {
	return t.mapper
}

func (t *_trigger) GetCluster() string {
	return t.cluster
}

func (t *_trigger) Error() error {
	return t.err
}

// enqueueRequestFromMapFuncFactory enqueue requests for effective clusters
func enqueueRequestFromMapFuncFactory(fn MapFuncFactory) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster sigcluster.Cluster) mchandler.EventHandler {
		return handler.TypedEnqueueRequestsFromMapFunc[client.Object, mcreconcile.Request](fn(clusterName, cluster))
	}
}
