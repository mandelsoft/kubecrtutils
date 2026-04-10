package controllers

import (
	"github.com/mandelsoft/kubecrtutils/cluster"
)

// --- begin settings ---
type Settings struct {
	Source cluster.ClusterEquivalent
	Target cluster.Cluster

	// common state
	Mapping Mapping
}

// --- end settings ---
