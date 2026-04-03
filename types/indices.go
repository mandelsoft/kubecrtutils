package types

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Index interface {
	GetName() string
	GetCluster() ClusterEquivalent
	GetGVK() schema.GroupVersionKind

	GetList(ctx context.Context, namespace, key string) (client.ObjectList, error)

	ForEachItem(ctx context.Context, namespace, key string, action func(ctx context.Context, obj runtime.Object) error) error
	Trigger(ctx context.Context, namespace, key string) error

	GetEffective() Index
	GetResource() client.Object
}

type Indices interface {
	internal.Group[Index]
}

type IndexerFactory = ClustersAware[client.IndexerFunc]

type IndexDefinition interface {
	GetName() string
	GetTarget() string
	GetResource() client.Object
	GetIndexer() IndexerFactory
	GetEffective() IndexDefinition
	ApplyMappings(mappings ControllerMappings) IndexDefinition
	Apply(ctx context.Context, set Clusters, logger logging.Logger) (Index, error)
}

type IndexDefinitions interface {
	internal.Definitions[IndexDefinition, IndexDefinitions]
	ApplyMappings(mappings ControllerMappings) IndexDefinitions

	GetIndices(ctx context.Context, clusters Clusters, logger logging.Logger) (Indices, error)
}
