package controller

import (
	"context"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/sliceutils"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type MapperFactory = func(ctx context.Context, cntr types.Controller) (myhandler.TypedMapFuncFactory[client.Object, mcreconcile.Request], error)

type TypedMapperFactory[T client.Object, R comparable] = func(ctx context.Context, cntr Controller) handler.TypedMapFunc[T, R]
type LocalTypedMapperFactory[T client.Object] TypedMapperFactory[T, reconcile.Request]

type ResourceTriggerDefinition interface {
	OnCluster(name string) ResourceTriggerDefinition

	GetDescription() string
	GetResource() client.Object
	GetMapper() MapperFactory
	GetCluster() string
	Error() error
}

type _trigger struct {
	desc    string
	proto   client.Object
	mapper  MapperFactory
	cluster string
	err     error
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
	return func(ctx context.Context, cntr types.Controller) (myhandler.MapFuncFactory, error) {
		return m, nil
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

func (t *_trigger) Error() error {
	return t.err
}

type Converter[I, O any] = func(I) O

type RequestConverterForCluster[R comparable] = func(clusterName string) Converter[R, mcreconcile.Request]

func ConvertMapFunc[O client.Object, R comparable](mapFunc handler.TypedMapFunc[O, R]) handler.TypedMapFunc[client.Object, R] {
	return func(ctx context.Context, object client.Object) []R {
		return mapFunc(ctx, any(object).(O))
	}
}

func LiftRequest(clusterName string) Converter[reconcile.Request, mcreconcile.Request] {
	return func(request reconcile.Request) mcreconcile.Request {
		return mcreconcile.Request{
			ClusterName: clusterName,
			Request:     request,
		}
	}
}

func CompleteRequest[R mcreconcile.ClusterAware[R]](clusterName string) Converter[R, R] {
	return func(request R) R {
		if request.Cluster() != "" {
			return request
		}
		return request.WithCluster(clusterName)
	}
}

func mapperFactoryForTypedFactory[T client.Object, R comparable](fac TypedMapperFactory[T, R], converter RequestConverterForCluster[R]) MapperFactory {
	return func(ctx context.Context, cntr Controller) (myhandler.TypedMapFuncFactory[client.Object, mcreconcile.Request], error) {
		m := fac(ctx, cntr)
		return func(clusterName string, _ sigcluster.Cluster) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
			conv := converter(clusterName)
			return func(ctx context.Context, obj client.Object) []mcreconcile.Request {
				return sliceutils.Transform(m(ctx, (any(obj)).(T)), conv)
			}
		}, nil
	}
}
