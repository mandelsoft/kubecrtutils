package cluster

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
)

////////////////////////////////////////////////////////////////////////////////

type _clusters struct {
	internal.Group[ClusterEquivalent]
	multi    bool
	disabled set.Set[string]
	mappings mapping.Mappings
}

var _ Clusters = (*_clusters)(nil)

func NewClusters() Clusters {
	return newClusters()
}

func newClusters() *_clusters {
	return &_clusters{Group: internal.NewGroup[ClusterEquivalent]("cluster"), disabled: set.New[string](), mappings: mapping.Mappings{}}
}

func (c *_clusters) GetMappings() mapping.Mappings {
	return c.mappings
}

func (c *_clusters) Get(name string) ClusterEquivalent {
	b, n := fpi.Split(name)
	if b == "" {
		return c.Group.Get(n)
	}
	f := c.Group.Get(b)
	if f == nil {
		return nil
	}
	if f.AsFleet() == nil {
		return nil
	}
	return f.AsFleet().GetClusterByLocalName(n)
}

func (c *_clusters) IsDisabled(name string) bool {
	return c.disabled.Has(name)
}

func (c *_clusters) GetClusterById(id string) ClusterEquivalent {
	for _, cl := range c.Elements {
		if cl.GetId() == id {
			return cl
		}
		fl := cl.AsFleet()
		if fl != nil {
			cl := fl.GetClusterById(id)
			if cl != nil {
				return cl
			}
		}
	}
	return nil
}

func (c *_clusters) Add(elems ...ClusterEquivalent) error {
	err := c.Group.Add(elems...)
	if err != nil {
		return err
	}
	for _, elem := range elems {
		if elem.AsFleet() != nil {
			c.multi = true
		}
		if c.Len() > 1 {
			c.multi = true
		}
		if elem.GetName() != elem.GetEffective().GetName() {
			c.mappings.Add(elem.GetName(), elem.GetEffective().GetName())
		}
	}
	return nil
}

func (c *_clusters) IsMulti() bool {
	return c.multi
}
