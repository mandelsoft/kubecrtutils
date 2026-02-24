package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/transformer"
	"github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/controller/helper"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apimtypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type IndexKeyTransformer[R comparable] = transformer.Transformer[R, string]

func DefaultRequestToIndexKey(req mcreconcile.Request) string {
	if req.ClusterName == "" {
		return req.Request.String()
	}
	return req.ClusterName + string(apimtypes.Separator) + req.Request.String()
}

func DefaultRequestToLocalIndexKey(req mcreconcile.Request) string {
	return req.Request.String()
}

func DefaultLocalRequestToIndexKey(req reconcile.Request) string {
	return req.String()
}

type ObjectToIndexKeyMapper[T client.Object] = ObjectMapper[T, string]
type ObjectToIndexKeyMapperFactory[T client.Object] = ControllerAware[ClusterAware[ObjectToIndexKeyMapper[T]]]

// //////////////////////////////////////////////////////////////////////////////
// OwnerIndexTriggers
// OwnerIndexTrigger maps a resource of type T to its owner of type O, which is then used
// as key for an index to trigger a reconcilation.
// The index key is composed by default with DefaultRequestToIndexKey
// incorporating the cluster name.
//
// Scenario:
//
//		C: Controller Resource
//		O: Owner Resource
//		W: Watch Resource
//
//		+---index---+
//		v           |
//		C ---ref--> O
//		            ^
//	             |
//	           owner
//		            |
//		            W
//
// An object C references some object O (of potentially any type).
// An index maintains maps the identity of O (including type) to the referencing
// objects C. The controller of O maintains an object of type W, which
// is watched. All owners are tried to map to C objects using the index.
// For those objects the reconcilation is triggered.
func OwnerIndexTrigger[T client.Object](name string, converter ...IndexKeyTransformer[owner.Owner]) ResourceTriggerDefinition {
	c := general.OptionalDefaulted[IndexKeyTransformer[owner.Owner]](owner.Owner.AsKey, converter...)
	o := mapOwnersFactory[T](false)
	m := GenericIndexKeyMapperByFactory(o, c)
	return IndexTrigger[T](name, m)
}

////////////////////////////////////////////////////////////////////////////////

func IndexKeyMapper[T client.Object](mapFunc handler.TypedMapFunc[T, mcreconcile.Request], converter ...IndexKeyTransformer[mcreconcile.Request]) ObjectToIndexKeyMapperFactory[T] {
	c := general.OptionalDefaulted[IndexKeyTransformer[mcreconcile.Request]](DefaultRequestToIndexKey, converter...)
	return GenericIndexKeyMapperForMapFunc(mapFunc, c)
}

func LocalIndexKeyMapper[T client.Object](mapFunc handler.TypedMapFunc[T, reconcile.Request], converter ...IndexKeyTransformer[reconcile.Request]) ObjectToIndexKeyMapperFactory[T] {
	c := general.OptionalDefaulted[IndexKeyTransformer[reconcile.Request]](DefaultLocalRequestToIndexKey, converter...)
	return GenericIndexKeyMapperForMapFunc(mapFunc, c)
}

func GenericIndexKeyMapperForMapFunc[T client.Object, R comparable](mapFunc handler.TypedMapFunc[T, R], converter transformer.Transformer[R, string]) ObjectToIndexKeyMapperFactory[T] {
	return handler.Lift(func(ctx context.Context, obj T) []string {
		return sliceutils.Transform[[]R, R, string](mapFunc(ctx, obj), converter)
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func IndexKeyMapperByFactory[T client.Object](mapFunc handler.TypedControllerAwareMapFuncFactory[T, mcreconcile.Request]) ObjectToIndexKeyMapperFactory[T] {
	return GenericIndexKeyMapperByFactory(mapFunc, DefaultRequestToIndexKey)
}

func LocalIndexKeyMapperByFactory[T client.Object](mapFunc handler.TypedControllerAwareMapFuncFactory[T, reconcile.Request]) ObjectToIndexKeyMapperFactory[T] {
	return GenericIndexKeyMapperByFactory(mapFunc, DefaultLocalRequestToIndexKey)
}

func GenericIndexKeyMapperByFactory[T client.Object, R comparable](mapFunc ControllerAware[ClusterAware[ObjectMapper[T, R]]], converter transformer.Transformer[R, string]) ObjectToIndexKeyMapperFactory[T] {
	return func(ctx context.Context, cntr types.Controller) (ClusterAware[ObjectToIndexKeyMapper[T]], error) {
		f, err := mapFunc(ctx, cntr)
		if err != nil {
			return nil, err
		}
		return func(clusterName string, cluster sigcluster.Cluster) ObjectToIndexKeyMapper[T] {
			m := f(clusterName, cluster)
			return func(ctx context.Context, obj T) []string {
				return sliceutils.Transform[[]R, R, string](m(ctx, obj), converter)
			}
		}, nil
	}
}

func transformMapper[T client.Object, R, O comparable, IF ControllerAware[ClusterAware[ObjectMapper[T, R]]], OF ControllerAware[ClusterAware[ObjectMapper[T, O]]]](in IF, converter transformer.Transformer[R, O]) OF {
	return func(ctx context.Context, cntr types.Controller) (ClusterAware[ObjectMapper[T, O]], error) {
		f, err := in(ctx, cntr)
		if err != nil {
			return nil, err
		}
		return func(clusterName string, cluster sigcluster.Cluster) ObjectMapper[T, O] {
			m := f(clusterName, cluster)
			return func(ctx context.Context, obj T) []O {
				return sliceutils.Transform[[]R, R, O](m(ctx, obj), converter)
			}
		}, nil
	}
}

////////////////////////////////////////////////////////////////////////////////

// IndexTrigger uses an index mapping a T object identity to the controller resource
// to trigger all objects on a changing T.
func IndexTrigger[T client.Object](indexName string, mapFunc ...ObjectToIndexKeyMapperFactory[T]) ResourceTriggerDefinition {
	return newTrigger[T](
		indexMappingFactory[T](indexName, mapFunc...),
		"index trigger "+indexName,
	)
}

func indexMappingFactory[T client.Object](indexName string, mapFunc ...ObjectToIndexKeyMapperFactory[T]) handler.ControllerAwareMapFuncFactory {
	return func(ctx context.Context, c types.Controller) (handler.MapFuncFactory, error) {
		idx := c.GetIndex(indexName)
		if idx == nil {
			return nil, fmt.Errorf("index %q not found", indexName)
		}
		var f ClusterAware[ObjectToIndexKeyMapper[T]]
		var err error
		if in := general.Optional(mapFunc...); in != nil {
			f, err = in(ctx, c)
			if err != nil {
				return nil, err
			}
		}
		log := c.GetLogger().WithName("indextrigger").WithName(indexName).WithValues("index", indexName, "gvk", idx.GetGVK())
		return func(clusterName string, cl sigcluster.Cluster) handler.MapFunc {
			var mapper ObjectToIndexKeyMapper[T]
			if f != nil {
				mapper = f(clusterName, cl)
			}
			conv := helper.LiftRequest(clusterName)
			return func(ctx context.Context, obj client.Object) []mcreconcile.Request {
				var keys []string
				var result []reconcile.Request
				key := client.ObjectKeyFromObject(obj)
				if mapper != nil {
					keys = mapper(ctx, obj.(T))
					log.Info(" {{object}} mapped to : {{mapped}} for index access", "object", key, "mapped", result, "cluster", clusterName)
				} else {
					keys = []string{key.String()}
				}
				for _, k := range keys {
					list := &unstructured.UnstructuredList{}
					list.SetGroupVersionKind(idx.GetGVK())
					err := cl.GetCache().List(ctx, list, client.InNamespace(obj.GetNamespace()), client.MatchingFields{indexName: k})
					if err != nil {
						return nil
					}
					r := sliceutils.Transform(list.Items, func(u unstructured.Unstructured) reconcile.Request {
						return reconcile.Request{client.ObjectKeyFromObject(&u)}
					})
					result = append(result, r...)
				}
				log.Info("trigger indexed for {{objects}}: {{triggered}}", "objects", keys, "triggered", result, "cluster", clusterName)
				return sliceutils.Transform(result, conv)
			}
		}, nil
	}
}
