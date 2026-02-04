package types

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/enqueue"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcctrl "sigs.k8s.io/multicluster-runtime"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type ControllerManager interface {
	GetName() string
	GetManager() mcctrl.Manager
	GetMainCluster() ClusterEquivalent
	GetCluster(name string) ClusterEquivalent
	GetClusters() Clusters
	GetIndex(name string) Index
	GetIndices() Indices

	GetLogger() logging.Logger
	GetControllerDefinition(name string) ControllerDefinition
}

type ControllerDefinition interface {
	flagutils.Options
	GetName() string
	GetCluster() string
	GetClusters() sets.Set[string]
	GetResource() client.Object
	GetWatchPredicates() []predicate.Predicate

	GetError() error
	GetOptions() flagutils.Options

	// CreateController handles the global definitions and provides
	// a Controller
	CreateController(ctx context.Context, mgr ControllerManager) (Controller, error)
}

type Controller interface {
	GetName() string
	GetFieldManager() string
	GetLogger() logging.Logger
	GetClusters() Clusters
	GetCluster() ClusterEquivalent
	GetResource() client.Object
	GetControllerManager() ControllerManager
	GetRecoder(ctx context.Context) record.EventRecorder
	GetReconciler() reconcile.Reconciler
	GetIndex(name string) Index

	Complete(ctx context.Context) error
}

type Controllers interface {
	internal.Group[Controller]
}

type Cluster interface {
	ClusterEquivalent
	enqueue.TypedMux[reconcile.Request]

	client.Client
	cluster.Cluster

	Unwrap() Cluster
	GetCluster() cluster.Cluster

	GetIndex(name string) Index

	GetTypeConverter() managedfields.TypeConverter
	ApplyTrigger(builder *ctrl.Builder, proto client.Object) error

	IsSameAs(ClusterEquivalent) bool
}

type Clusters interface {
	internal.Group[ClusterEquivalent]
}

type Index interface {
	GetName() string
	GetCluster() ClusterEquivalent
	GetList(ctx context.Context, namespace, key string) (client.ObjectList, error)

	ForEachItem(ctx context.Context, namespace, key string, action func(ctx context.Context, obj runtime.Object) error) error
	Trigger(ctx context.Context, namespace, key string) error
}

type Indices interface {
	internal.Group[Index]
}

type ClusterDefinitionProvider interface {
	GetDefinition() ClusterDefinition
}

type ClusterDefinition interface {
	ClusterDefinitionProvider
	flagutils.Options
	flagutils.OptionSetProvider

	RequireFleet() bool
	RequireIdentity()

	GetConfig(*config.ConfigOptions) (*config.Config, error)

	GetName() string
	GetFallback() string
	GetDescription() string
	GetScheme() *runtime.Scheme

	WithFallback(fallback string) ClusterDefinition
	WithScheme(scheme *runtime.Scheme) ClusterDefinition

	Create(defs ClusterDefinitions) (ClusterEquivalent, error)
}

type ClusterDefinitions interface {
	internal.Definitions[ClusterDefinition, ClusterDefinitions]
	flagutils.Validatable

	WithScheme(scheme *runtime.Scheme) ClusterDefinitions
	GetError() error

	GetClusters() Clusters
	GetScheme() *runtime.Scheme
}

type Fleet interface {
	ClusterEquivalent
	enqueue.TypedMux[mcreconcile.Request]

	GetType() FleetType

	GetProvider() multicluster.Provider
	GetClusterNames() []string
	GetCluster(name string) Cluster

	GetClusterById(id string) ClusterEquivalent

	GetBaseCluster() Cluster

	Strip(name string) string
	Compose(name string) string
}

type FleetType interface {
	GetType() string
	Create(defs ClusterDefinitions, definition ClusterDefinition, config config.Config) (Fleet, error)
	GetRules(def ClusterDefinition) config.Rules
}

type ClusterEquivalent interface {
	client.FieldIndexer
	enqueue.Mux

	GetName() string
	GetId() string
	GetInfo() string

	GetScheme() *runtime.Scheme
	GetEffective() ClusterEquivalent
	GetClusterById(id string) ClusterEquivalent

	CreateIndex(ctx context.Context, name string, proto client.Object, indexer client.IndexerFunc, wrap ...func(cluster ClusterEquivalent, name string) (Index, error)) (Index, error)

	// Get for a fleet uses the cluster provided by the context. If no fleet member is
	// provided an error is returned.
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error

	// List for a fleet uses the cluster provided by the context. If no fleet member is
	// provided an error is returned.
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error

	AsCluster() Cluster
	AsFleet() Fleet

	IsSameAs(o ClusterEquivalent) bool

	Filter(clusterName string, cluster cluster.Cluster) bool
	FilterById(clusterId string) bool
	Match(clusterName string) bool
}

type SchemeProvider interface {
	GetScheme() *runtime.Scheme
}

type ObjectModifier interface {
	Modify(cluster Cluster, obj client.Object) error
}

type ObjectModifierFunc func(cluster Cluster, obj client.Object) error

func (f ObjectModifierFunc) Modify(cluster Cluster, obj client.Object) error {
	return f(cluster, obj)
}
