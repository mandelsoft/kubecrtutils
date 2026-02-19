package controller

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/sliceutils"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

// IndexTrigger uses an index mapping a T object identity to the controller resource
// to trigger all objects on a changing T.
func IndexTrigger[T client.Object](indexName string) ResourceTriggerDefinition {
	return newTrigger[T](
		indexMappingFactory[T](indexName),
		"index trigger "+indexName,
	)
}

func indexMappingFactory[T client.Object](indexName string) MapperFactory {
	return func(ctx context.Context, c types.Controller) (myhandler.MapFuncFactory, error) {
		idx := c.GetIndex(indexName)
		if idx == nil {
			return nil, fmt.Errorf("index %q not found", indexName)
		}
		log := c.GetLogger().WithName("indextrigger").WithName(indexName).WithValues("index", indexName, "gkv", idx.GetGVK())
		return func(clusterName string, cl cluster.Cluster) handler.TypedMapFunc[client.Object, mcreconcile.Request] {
			conv := LiftRequest(clusterName)
			return func(ctx context.Context, obj client.Object) []mcreconcile.Request {
				key := client.ObjectKeyFromObject(obj).String()
				list := &unstructured.UnstructuredList{}
				list.SetGroupVersionKind(idx.GetGVK())
				err := cl.GetCache().List(ctx, list, client.InNamespace(obj.GetNamespace()), client.MatchingFields{indexName: key})
				if err != nil {
					return nil
				}
				result := sliceutils.Transform(list.Items, func(u unstructured.Unstructured) reconcile.Request {
					return reconcile.Request{client.ObjectKeyFromObject(&u)}
				})
				log.Info("trigger indexed for {{object}}: {{triggered}}", "object", key, "triggered", result, "cluster", clusterName)
				return sliceutils.Transform(result, conv)
			}
		}, nil
	}
}
