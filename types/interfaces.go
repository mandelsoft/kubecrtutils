package types

import (
	"context"

	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type NameSet = set.Set[string]

type DefinitionProvider[D any] interface {
	GetDefinition() D
}

type IndexerFunc[T client.Object] = func(T) []string

type ObjectMapper[T any, R any] = func(ctx context.Context, obj T) []R

type ControllerAware[T any] = func(ctx context.Context, cntr Controller) (T, error)
type ClusterAware[T any] = func(clusterName string, cluster cluster.Cluster) T
type ClustersAware[T any] = func(ctx context.Context, logger logging.Logger, clusters Clusters) (T, error)

type ClusterMatcher func(clusterId string) (clusterName string, equal bool)

type SchemeProvider interface {
	GetScheme() *runtime.Scheme
}

type ObjectModifier interface {
	Modify(cluster Cluster, obj client.Object) error
}

type ObjectModifierFunc func(cluster Cluster, obj client.Object) error

func (f ObjectModifierFunc) Modify(cluster Cluster, obj client.Object) error {
	return f(cluster, obj)
}
