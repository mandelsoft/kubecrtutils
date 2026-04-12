package types

import (
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"github.com/mandelsoft/logging"
)

type ComponentNames = plain.ComponentNames

type ComponentFilter interface {
	GetUsedComponents(ConstraintContext) ComponentNames
}

// --- begin component ---

type Component interface {
	logging.Logger

	GetName() string

	GetCluster(name string) ClusterEquivalent
	GetComponent(name string) Component
	GetIndex(name string) Index

	GetIndices() Indices
	GetClusters() Clusters

	GetImplementation() ComponentImplementation
}

// --- end component ----

// --- begin component implementation ---

type ComponentImplementation interface {
	GetComponent() Component
}

// --- end component implementation ---

type Components interface {
	internal.Group[Component]
	IsDisabled(name string) bool
	Map(mapping mapping.Mappings, names ComponentNames) (Components, error)
}
