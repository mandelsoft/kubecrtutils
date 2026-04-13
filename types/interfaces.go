package types

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/types/plain"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type NameSet = plain.NameSet

type DefinitionProvider[D any] interface {
	GetDefinition() D
}

type ObjectMapper[T any, R any] = func(ctx context.Context, obj T) []R

type ControllerAware[T any] = func(ctx context.Context, cntr Controller) (T, error)
type ClusterAware[T any] = func(clusterName string, cluster cluster.Cluster) T
type ClustersAware[T any] = func(ctx context.Context, logger logging.Logger, clusters Clusters) (T, error)

type SchemeProvider interface {
	GetScheme() *runtime.Scheme
}

func AsSchemeProvider(s *runtime.Scheme) SchemeProvider {
	return _schemeprovider{s}
}

type _schemeprovider struct {
	s *runtime.Scheme
}

func (p _schemeprovider) GetScheme() *runtime.Scheme {
	return p.s
}

type ObjectModifier interface {
	Modify(cluster Cluster, obj client.Object) error
}

type ObjectModifierFunc func(cluster Cluster, obj client.Object) error

func (f ObjectModifierFunc) Modify(cluster Cluster, obj client.Object) error {
	return f(cluster, obj)
}
