package controllers

import (
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller/replication"
)

// --- begin settings ---
type Settings struct {
	Source cluster.ClusterEquivalent
	Target cluster.Cluster

	// common state
	Mapping replication.ResourceMapping
}

// --- end settings ---
