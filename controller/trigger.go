package controller

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/controller/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type ResourceTriggerDefinition interface {
	OnCluster(name string) ResourceTriggerDefinition

	GetDescription() string
	GetResource() client.Object
	GetMapper() handler.ControllerAwareMapFuncFactory
	GetCluster() string
	Error() error
}

type _trigger struct {
	desc    string
	proto   client.Object
	mapper  handler.ControllerAwareMapFuncFactory
	cluster string
	err     error
}

func newTrigger[T client.Object](mapper handler.ControllerAwareMapFuncFactory, desc ...string) *_trigger {
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

func (t *_trigger) GetMapper() handler.ControllerAwareMapFuncFactory {
	return t.mapper
}

func (t *_trigger) GetCluster() string {
	return t.cluster
}

func (t *_trigger) Error() error {
	return t.err
}

// enqueueRequestFromMapFuncFactory enqueue requests for effective clusters
func enqueueRequestFromMapFuncFactory(fn handler.MapFuncFactory) mchandler.EventHandlerFunc {
	return func(clusterName string, cluster sigcluster.Cluster) mchandler.EventHandler {
		return handler.TypedEnqueueRequestsFromMapFunc[client.Object, mcreconcile.Request](fn(clusterName, cluster))
	}
}
