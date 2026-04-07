package types

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"github.com/mandelsoft/logging"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ControllerNames = plain.ControllerNames

type ControllerSet interface {
	GetNames() []string
	GetGroups() map[string][]string
}

type ControllerSource interface {
	GetControllerSet() ControllerSet
}

type ControllerDefinition interface {
	flagutils.Options

	GetName() string

	GetActivationConstraints() Constraints
	GetCluster() string
	GetClusters() ClusterNames
	GetComponents() ComponentNames
	GetResource() client.Object
	GetGroups() NameSet
	GetWatchPredicates() []predicate.Predicate
	GetFinalizer() string

	GetForeignIndices() IndexDefinitions

	GetRequiredClusters(mappings mapping.ControllerMappings) ClusterNames
	GetRequiredComponents(mappings mapping.ControllerMappings) ComponentNames

	GetError() error
	GetOptions() flagutils.Options

	IndexProvider
	Applyable
}

type ControllerDefinitions interface {
	internal.Definitions[ControllerDefinition, ControllerDefinitions]
	AddRule(...Constraint) ControllerDefinitions

	flagutils.Validatable

	ControllerSource
	ClusterFilter
	ComponentFilter

	IndexProvider
	Applyable
}

type Controller interface {
	GetName() string
	GetOptions() flagutils.Options
	GetFieldManager() string
	GetFinalizer() string
	GetLogger() logging.Logger
	GetClusterMappings() mapping.Mappings
	GetClusters() Clusters
	GetComponents() Components
	GetIndices() Indices
	GetCluster() ClusterEquivalent
	GetLogicalCluster(name string) ClusterEquivalent
	GetResource() client.Object
	GetControllerManager() ControllerManager
	GetRecoder(ctx context.Context) record.EventRecorder
	GetReconciler() reconcile.Reconciler
	GetIndex(name string) Index
	GetOwnerHandler() OwnerHandler

	Complete(ctx context.Context) error

	GenerateNameFor(ctx context.Context, tgt Cluster, prefix, namespace, name string, len ...int) string
}

type Controllers interface {
	internal.Group[Controller]
}
