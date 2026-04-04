package component

import (
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/logging"
)

type Base struct {
	logging.Logger

	def Definition

	self     Component
	clusters cluster.Clusters
	comps    Components
	indices  cacheindex.Indices
}

func (c *Base) GetName() string {
	return c.def.GetName()
}

func (c *Base) GetDefinition() Definition {
	return c.def
}

func (c *Base) GetEffective() Component {
	return c.self
}

func (c *Base) GetCluster(name string) cluster.ClusterEquivalent {
	return c.clusters.Get(name)
}

func (c *Base) GetComponent(name string) Component {
	return c.comps.Get(name)
}

func (c *Base) GetIndex(name string) cacheindex.Index {
	return c.indices.Get(name)
}

func (c *Base) GetIndices() cacheindex.Indices {
	return c.indices
}
