package types

import (
	"github.com/mandelsoft/logging"
	mcctrl "sigs.k8s.io/multicluster-runtime"
)

type ControllerManager interface {
	GetName() string
	GetManager() mcctrl.Manager
	GetMainCluster() ClusterEquivalent
	MapTechnicalName(name string) ClusterEquivalent
	GetComponents() Components
	GetClusters() Clusters
	GetIndex(name string) Index
	GetIndices() Indices

	GetLogger() logging.Logger
	GetControllerDefinition(name string) ControllerDefinition
}
