package component

import (
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
)

type Component = types.Component
type ComponentImplementation = types.ComponentImplementation

type _component struct {
	logging.Logger

	def Definition

	impl     ComponentImplementation
	clusters cluster.Clusters
	comps    Components
	indices  cacheindex.Indices
}

func (c *_component) GetName() string {
	return c.def.GetName()
}

func (c *_component) GetDefinition() Definition {
	return c.def
}

func (c *_component) GetImplementation() ComponentImplementation {
	return c.impl
}

func (c *_component) GetCluster(name string) cluster.ClusterEquivalent {
	return c.clusters.Get(name)
}

func (c *_component) GetComponent(name string) Component {
	return c.comps.Get(name)
}

func (c *_component) GetIndex(name string) cacheindex.Index {
	return c.indices.Get(name)
}

func (c *_component) GetIndices() cacheindex.Indices {
	return c.indices
}

func (c *_component) GetClusters() cluster.Clusters {
	return c.clusters
}
