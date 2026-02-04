package controller

import (
	"context"
	"reflect"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/controller/helper"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type MapperFactory interface {
	MultiTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, mcreconcile.Request]
	SingleTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, reconcile.Request]
}

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

func newTrigger[T any, P kubecrtutils.ObjectPointer[T]](mapper MapperFactory, desc ...string) *_trigger {
	var resource T
	return &_trigger{
		desc:   general.OptionalDefaulted("resource mapping", desc...),
		proto:  any(&resource).(P),
		mapper: mapper,
	}
}

////////////////////////////////////////////////////////////////////////////////

func ResourceTrigger[T any, P kubecrtutils.ObjectPointer[T]](mapFunc handler.TypedMapFunc[P, mcreconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTrigger[T, P](
		multiMapperFactory(ConvertMapFunc[P, mcreconcile.Request](mapFunc)),
		desc...)
}

func multiMapperFactory(fn myhandler.MapFunc) MapperFactory {
	return &multiFactory{fn}
}

type multiFactory struct {
	mapper myhandler.MapFunc
}

func (f *multiFactory) MultiTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
	if cntr.GetCluster().Match(current.GetName()) {
		return withClusterCompletion(current.GetName(), f.mapper)
	}
	return f.mapper
}

func (f *multiFactory) SingleTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, reconcile.Request] {
	return helper.FilterCluster(cntr.GetCluster().Match, f.mapper)
}

////////////////////////////////////////////////////////////////////////////////

func LocalResourceTrigger[T any, P kubecrtutils.ObjectPointer[T]](mapFunc handler.TypedMapFunc[P, reconcile.Request], desc ...string) ResourceTriggerDefinition {
	return newTrigger[T, P](
		localMapperFactory(ConvertMapFunc[P, reconcile.Request](mapFunc)),
		desc...)
}

func localMapperFactory(fn handler.MapFunc) MapperFactory {
	return &localFactory{fn}
}

type localFactory struct {
	mapper handler.MapFunc
}

func (f *localFactory) MultiTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
	if cntr.GetCluster().Match(current.GetName()) {
		return withClusterInjection(current.GetName(), f.mapper)
	}
	// Ooops, this looks like an invalid combi.
	return nil
}

func (f *localFactory) SingleTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, reconcile.Request] {
	return f.mapper
}

////////////////////////////////////////////////////////////////////////////////

func OwnerTrigger[T any, P kubecrtutils.ObjectPointer[T]]() ResourceTriggerDefinition {
	var obj T

	return newTrigger[T, P](
		&multiOwnerFactory{any(&obj).(P)},
		"owner trigger",
	)
}

type multiOwnerFactory struct {
	proto client.Object
}

func (f *multiOwnerFactory) MultiTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
	return owner.MapOwnerToRequestByObject(owner.NewHandler(cntr.GetCluster().GetScheme()), owner.MatcherFor(cntr.GetCluster()), cntr.GetCluster(), current, f.proto)
}

func (f *multiOwnerFactory) SingleTarget(current types.Cluster, cntr types.Controller) handler.TypedMapFunc[client.Object, reconcile.Request] {
	return helper.FilterCluster(nil, f.MultiTarget(current, cntr))
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

func objectOf[T any]() T {
	t := generics.TypeOf[T]()
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem()).Interface().(T)
	}
	return reflect.New(t).Elem().Interface().(T)
}
