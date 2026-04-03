package cluster

import (
	"github.com/mandelsoft/kubecrtutils/types"
)

type SchemeProvider = types.SchemeProvider

type ClusterEquivalent = types.ClusterEquivalent

type Fleet = types.Fleet
type Cluster = types.Cluster
type Clusters = types.Clusters
type ClusterNames = types.ClusterNames
type Index = types.Index

type ClusterFilter = types.ClusterFilter

func GetClusterFor(c ClusterEquivalent, name string) Cluster {
	if c == nil {
		return nil
	}
	if f := c.AsFleet(); f != nil {
		return f.GetCluster(name)
	}
	if f := c.AsCluster(); f != nil {
		if f.GetName() == name {
			return f
		}
	}
	return nil
}
