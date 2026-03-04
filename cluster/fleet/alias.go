package fleet

import (
	"context"
	"fmt"
	"sync"

	"github.com/mandelsoft/kubecrtutils/cluster/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type _fleetAlias struct {
	lock sync.Mutex
	Fleet
	fpi.Composer

	clusters map[string]types.Cluster
}

func NewAlias(name string, c Fleet) Fleet {
	if c.GetName() == name {
		return c
	}
	a := &_fleetAlias{Composer: fpi.NewComposer(name), Fleet: c, clusters: make(map[string]types.Cluster)}
	return a
}

func (c *_fleetAlias) Compose(name string) string {
	return c.Composer.Compose(name)
}

func (c *_fleetAlias) Match(name string) bool {
	return c.Composer.Match(name) || c.Fleet.Match(name)
}

func (c *_fleetAlias) GetName() string {
	return c.Composer.GetName()
}

func (c *_fleetAlias) Unwrap() Fleet {
	return c.Fleet
}

func (c *_fleetAlias) GetClusterByLocalName(name string) types.Cluster {
	var f types.Cluster

	c.lock.Lock()
	defer c.lock.Unlock()
	n := c.Fleet.GetClusterByLocalName(name)
	if n != nil {
		f = c.clusters[name]
		if f == nil || f.GetEffective() != n {
			f = cluster.NewAlias(name, n)
			c.clusters[name] = f
		}
	} else {
		delete(c.clusters, name)
	}
	return f
}

func (c *_fleetAlias) GetCluster(name string) types.Cluster {
	b, n := fpi.Split(name)
	if c.GetName() == b {
		return c.GetClusterByLocalName(n)
	}
	return nil
}

func (a *_fleetAlias) LiftTechnical(name string) (string, types.Cluster) {
	b, n := fpi.Split(name)
	if a.GetName() == b {
		return a.Compose(n), a.GetClusterByLocalName(n)
	}
	panic(fmt.Errorf("technical cluster %q does not match logical fleet %q[%s]", name, a.GetName(), a.GetEffective().GetName()))
}

func (a *_fleetAlias) mapNames(list []types.GlobalKey, err error) ([]types.GlobalKey, error) {
	if err != nil {
		return nil, err
	}
	eff := a.GetEffective().GetName()
	for i := range list {
		b, n := fpi.Split(list[i].ClusterName)
		if b == eff {
			list[i].ClusterName = a.Compose(n)
		}
	}
	return list, err
}

func (a *_fleetAlias) ListIndexedGlobalKeys(ctx context.Context, obj runtime.Object, index string, key string, opts ...client.ListOption) ([]types.GlobalKey, error) {
	return a.mapNames(a.GetEffective().ListIndexedGlobalKeys(ctx, obj, index, key, opts...))
}

func (a *_fleetAlias) ListIndexedGlobalKeysByObjectKey(ctx context.Context, obj runtime.Object, index string, key types.TypedGlobalKey, opts ...client.ListOption) ([]types.GlobalKey, error) {
	return a.mapNames(a.GetEffective().ListIndexedGlobalKeysByObjectKey(ctx, obj, index, key, opts...))
}
