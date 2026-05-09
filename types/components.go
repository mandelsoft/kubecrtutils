package types

import (
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"github.com/mandelsoft/logging"
)

type ComponentNames = plain.ComponentNames

type ComponentFilter interface {
	GetUsedComponents(ConstraintContext) ComponentNames
}

type ComponentDefinition = interface {
	GetError() error

	IndexProvider
	Applyable

	flagutils.Options
	mapping.Consumer
	HealthDefinition[Component]

	GetName() string
	GetOptions() flagutils.Options

	GetActivationConstraints() Constraints

	GetForeignIndices() IndexDefinitions
	GetRequiredClusters(mappings mapping.ControllerMappings) ClusterNames
	GetRequiredComponents(mappings mapping.ControllerMappings) ComponentNames
}

// --- begin component ---

type Component interface {
	GetName() string
	GetLogger() logging.Logger
	GetOptions() flagutils.Options
	GetClusterMappings() mapping.Mappings

	GetControllerManager() ControllerManager
	GetIndices() Indices
	GetIndex(name string) Index
	GetClusters() Clusters
	GetCluster(name string) ClusterEquivalent
	GetComponents() Components
	GetComponent(name string) Component

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
