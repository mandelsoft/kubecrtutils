package cluster

import (
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
	"github.com/mandelsoft/kubecrtutils/internal"
)

////////////////////////////////////////////////////////////////////////////////

type clusters struct {
	internal.Group[ClusterEquivalent]
}

var _ Clusters = (*clusters)(nil)

func NewClusters() Clusters {
	return &clusters{internal.NewGroup[ClusterEquivalent]("cluster")}
}

func (c *clusters) Get(name string) ClusterEquivalent {
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

func (c *clusters) GetClusterById(id string) ClusterEquivalent {
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
