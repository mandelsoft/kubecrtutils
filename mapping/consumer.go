package mapping

import (
	"maps"

	"github.com/mandelsoft/kubecrtutils/types/plain"
)

type DefaultClusterConsumer struct {
	clusters plain.ClusterNames
}

var _ Consumer = (*DefaultClusterConsumer)(nil)

func NewDefaultClusterConsumer(names ...string) *DefaultClusterConsumer {
	return &DefaultClusterConsumer{
		clusters: plain.NewNameSet(names...),
	}
}

func (d *DefaultClusterConsumer) GetComponents() plain.ComponentNames {
	return plain.NewNameSet()
}

func (d *DefaultClusterConsumer) GetRequiredComponents(_ ControllerMappings) plain.ComponentNames {
	return plain.NewNameSet()
}

func (d *DefaultClusterConsumer) GetClusters() plain.ClusterNames {
	return maps.Clone(d.clusters)
}

func (d *DefaultClusterConsumer) GetRequiredClusters(mappings ControllerMappings) plain.ClusterNames {
	return DefaultMappings(mappings).ClusterMappings().GetMapped(d.GetClusters())
}

////////////////////////////////////////////////////////////////////////////////

type DefaultConsumer struct {
	DefaultClusterConsumer
	components plain.ComponentNames
}

var _ Consumer = (*DefaultConsumer)(nil)

func NewDefaultConsumer() *DefaultConsumer {
	return &DefaultConsumer{
		DefaultClusterConsumer: *NewDefaultClusterConsumer(),
		components:             plain.ComponentNames{},
	}
}

func (d *DefaultConsumer) UseCluster(name ...string) {
	d.clusters.Add(name...)
}

func (d *DefaultConsumer) UseComponent(name ...string) {
	d.components.Add(name...)

}

func (d *DefaultConsumer) GetComponents() plain.ComponentNames {
	return maps.Clone(d.components)
}

func (d *DefaultConsumer) GetRequiredComponents(mappings ControllerMappings) plain.ComponentNames {
	return DefaultMappings(mappings).ComponentMappings().GetMapped(d.GetComponents())
}
