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
	GetClusters() Clusters
	GetIndices() Indices
	GetIndex(name string) Index
	GetComponents() Components
	GetControllers() Controllers
	GetOwnerHandler(scheme SchemeProvider) OwnerHandler

	GetLogger() logging.Logger
	GetControllerDefinition(name string) ControllerDefinition
}
