package cluster

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
	"github.com/mandelsoft/kubecrtutils/internal"
)

////////////////////////////////////////////////////////////////////////////////

type clusters struct {
	internal.Group[ClusterEquivalent]
	multi    bool
	disabled set.Set[string]
}

var _ Clusters = (*clusters)(nil)

func NewClusters() Clusters {
	return newClusters()
}

func newClusters() *clusters {
	return &clusters{Group: internal.NewGroup[ClusterEquivalent]("cluster"), disabled: set.New[string]()}
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

func (c *clusters) IsDisabled(name string) bool {
	return c.disabled.Has(name)
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

func (c *clusters) Add(elems ...ClusterEquivalent) error {
	err := c.Group.Add(elems...)
	if err != nil {
		return err
	}
	if !c.multi {
		for _, cl := range elems {
			if cl.AsFleet() != nil {
				c.multi = true
			}
			if c.Len() > 1 {
				c.multi = true
			}
		}
	}
	return nil
}

func (c *clusters) IsMulti() bool {
	return c.multi
}
