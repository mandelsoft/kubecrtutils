package cluster

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type _clusterAlias struct {
	Cluster
	name string
}

func NewAlias(name string, c Cluster) Cluster {
	if name == c.GetName() {
		return c
	}
	return &_clusterAlias{
		Cluster: c,
		name:    name,
	}
}

func (a *_clusterAlias) GetName() string {
	return a.name
}

func (a *_clusterAlias) Unwrap() Cluster {
	return a.Cluster
}

func (a *_clusterAlias) LiftTechnical(clusterName string) (string, Cluster) {
	a.Cluster.LiftTechnical(clusterName) // panic to indicate  corruption
	return a.name, a
}

func (a *_clusterAlias) mapNames(list []types.GlobalKey, err error) ([]types.GlobalKey, error) {
	if err != nil {
		return nil, err
	}
	eff := a.GetEffective().GetName()
	for i := range list {
		if list[i].ClusterName == eff {
			list[i].ClusterName = a.GetName()
		}
	}
	return list, err
}

func (a *_clusterAlias) ListIndexedGlobalKeys(ctx context.Context, obj runtime.Object, index string, key string, opts ...client.ListOption) ([]types.GlobalKey, error) {
	return a.mapNames(a.GetEffective().ListIndexedGlobalKeys(ctx, obj, index, key, opts...))
}

func (a *_clusterAlias) ListIndexedGlobalKeysByObjectKey(ctx context.Context, obj runtime.Object, index string, key types.TypedGlobalKey, opts ...client.ListOption) ([]types.GlobalKey, error) {
	return a.mapNames(a.GetEffective().ListIndexedGlobalKeysByObjectKey(ctx, obj, index, key, opts...))
}
