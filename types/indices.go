package types

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// --- begin indexer func ---

type IndexerFunc[T client.Object] = func(T) []string

// --- end indexer func ---

type IndexNames = plain.IndexNames

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

type IndexSource interface {
	GetIndex(name string) Index
}

type Indices interface {
	internal.Group[Index]
}

type IndexerFactory = ClustersAware[client.IndexerFunc]

type IndexDefinition interface {
	mapping.Consumer
	GetName() string
	GetTarget() string
	GetResource() client.Object
	GetIndexer() IndexerFactory
	GetEffective() IndexDefinition
	Applyable
}

type IndexDefinitions interface {
	internal.Definitions[IndexDefinition, IndexDefinitions]
	IndexProvider
	// don't have other elements, therefore, not yet Applyable.
}
