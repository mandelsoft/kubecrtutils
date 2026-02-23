package cluster

import (
	"github.com/mandelsoft/kubecrtutils/cluster/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet"
)

func NewAlias(name string, c ClusterEquivalent) ClusterEquivalent {
	if c.GetName() == name {
		return c
	}
	if f := c.AsFleet(); f != nil {
		return fleet.NewAlias(name, f)
	}
	return cluster.NewAlias(name, c.AsCluster())
}
