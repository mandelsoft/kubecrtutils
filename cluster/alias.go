package cluster

import (
	"sync"

	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
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

func (c *_clusterAlias) GetName() string {
	return c.name
}

func (c *_clusterAlias) Unwrap() Cluster {
	return c.Cluster
}

////////////////////////////////////////////////////////////////////////////////

type _fleetAlias struct {
	lock sync.Mutex
	Fleet
	fpi.Composer

	clusters map[string]Cluster
}

func NewFleetAlias(name string, c Fleet) Fleet {
	if c.GetName() == name {
		return c
	}
	a := &_fleetAlias{Composer: fpi.NewComposer(name), Fleet: c, clusters: make(map[string]Cluster)}
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

func (c *_fleetAlias) GetCluster(name string) Cluster {
	var f Cluster
	base, l := fpi.Split(name)
	if c.GetName() != base {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	n := c.Fleet.GetCluster(c.Fleet.Compose(l))
	if n != nil {
		f = c.clusters[base]
		if f == nil || f.GetCluster() != f {
			f = NewAlias(name, n)
			c.clusters[base] = f
		}
	} else {
		delete(c.clusters, base)
	}
	return f
}

////////////////////////////////////////////////////////////////////////////////

type _clusterLikeAlias struct {
	lock sync.Mutex
	ClusterEquivalent
	cluster Cluster
	fleet   Fleet
	name    string
}

func NewClusterLikeAlias(name string, c ClusterEquivalent) ClusterEquivalent {
	return &_clusterLikeAlias{
		ClusterEquivalent: c,
		name:              name,
	}
}

func (c *_clusterLikeAlias) GetName() string {
	return c.name
}

func (c *_clusterLikeAlias) GetId() string {
	return c.ClusterEquivalent.GetId()
}

func (c *_clusterLikeAlias) GetEffective() ClusterEquivalent {
	return c.ClusterEquivalent.GetEffective()
}

func (c *_clusterLikeAlias) Unwrap() ClusterEquivalent {
	return c.ClusterEquivalent
}

func (c *_clusterLikeAlias) AsCluster() Cluster {
	c.lock.Lock()
	if c.cluster == nil {
		n := c.ClusterEquivalent.AsCluster()
		if n != nil {
			c.cluster = NewAlias(c.name, n)
		}
	}
	c.lock.Unlock()
	return c.cluster
}

func (c *_clusterLikeAlias) AsFleet() Fleet {
	c.lock.Lock()
	if c.fleet == nil {
		n := c.ClusterEquivalent.AsFleet()
		if n != nil {
			c.fleet = NewFleetAlias(c.name, n)
		}
	}
	c.lock.Unlock()
	return c.fleet

}
