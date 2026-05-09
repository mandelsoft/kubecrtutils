package component

import (
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
)

type Component = types.Component
type Implementation = types.ComponentImplementation

type _component struct {
	def        Definition
	logger     logging.Logger
	mappings   mapping.Mappings
	manager    types.ControllerManager
	components Components
	indices    cacheindex.Indices

	clusters cluster.Clusters
	tname    string
	impl     Implementation
}

func (c *_component) GetName() string {
	return c.tname
}

func (c *_component) GetControllerManager() types.ControllerManager {
	return c.manager
}

func (c *_component) GetLogger() logging.Logger {
	return c.logger
}

func (c *_component) GetClusterMappings() mapping.Mappings {
	return c.mappings
}

func (c *_component) GetDefinition() Definition {
	return c.def
}

func (c *_component) GetImplementation() Implementation {
	return c.impl
}

func (c *_component) GetCluster(name string) cluster.ClusterEquivalent {
	return c.clusters.Get(name)
}

func (c *_component) GetOptions() flagutils.Options {
	return c.def.GetOptions()
}

func (c *_component) GetComponent(name string) Component {
	return c.components.Get(name)
}

func (c *_component) GetComponents() Components {
	return c.components
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
