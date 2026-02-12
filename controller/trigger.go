package controller

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type MapperFactory = func(ctx context.Context, cntr types.Controller) myhandler.MapFuncFactory

type ResourceTriggerDefinition interface {
	OnCluster(name string) ResourceTriggerDefinition

	GetDescription() string
	GetResource() client.Object
	GetMapper() MapperFactory
	GetCluster() string
}

type _trigger struct {
	desc    string
	proto   client.Object
	mapper  MapperFactory
	cluster string
}

func newTriggerF[T client.Object](mapper myhandler.MapFuncFactory, desc ...string) *_trigger {
	return newTrigger[T](defaultMapperFactory(mapper), desc...)
}

func newTrigger[T client.Object](mapper MapperFactory, desc ...string) *_trigger {
	return &_trigger{
		desc:   general.OptionalDefaulted("resource mapping", desc...),
		proto:  generics.ObjectFor[T](),
		mapper: mapper,
	}
}

func defaultMapperFactory(m myhandler.MapFuncFactory) MapperFactory {
	return func(ctx context.Context, cntr types.Controller) myhandler.MapFuncFactory {
		return m
	}
}

////////////////////////////////////////////////////////////////////////////////

func ResourceTrigger[T client.Object](mapFunc handler.TypedMapFunc[T, mcreconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTriggerF[T](
		myhandler.MapFuncFactoryWithClusterCompletion(ConvertMapFunc[T, mcreconcile.Request](mapFunc)),
		desc...)
}

////////////////////////////////////////////////////////////////////////////////

func LocalResourceTrigger[T client.Object](mapFunc handler.TypedMapFunc[T, reconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTriggerF[T](
		myhandler.MapFuncFactoryWithClusterInjection(ConvertMapFunc[T, reconcile.Request](mapFunc)),
		desc...)
}

////////////////////////////////////////////////////////////////////////////////

func OwnerTrigger[T client.Object]() ResourceTriggerDefinition {
	return newTrigger[T](
		ownerFunc[T],
		"owner trigger",
	)
}

func ownerFunc[T client.Object](ctx context.Context, c types.Controller) myhandler.MapFuncFactory {
	return func(clusterName string, _ cluster.Cluster) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
		cl := c.GetControllerManager().MapTechnicalName(clusterName)
		return owner.MapOwnerToRequestByObject(owner.NewHandler(cl), owner.MatcherFor(c.GetCluster()), c.GetCluster(), cl.AsCluster(), c.GetResource())
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

func (t *_trigger) GetMapper() MapperFactory {
	return t.mapper
}

func (t *_trigger) GetCluster() string {
	return t.cluster
}

func ConvertMapFunc[O client.Object, R comparable](mapFunc handler.TypedMapFunc[O, R]) handler.TypedMapFunc[client.Object, R] {
	return func(ctx context.Context, object client.Object) []R {
		return mapFunc(ctx, any(object).(O))
	}
}
