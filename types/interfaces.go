package types

import (
	"context"
	"net/url"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/enqueue"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/merge"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcctrl "sigs.k8s.io/multicluster-runtime"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type IndexerFunc[T client.Object] = func(T) []string

type ObjectMapper[T any, R any] = func(ctx context.Context, obj T) []R
type ControllerAware[T any] func(ctx context.Context, cntr Controller) (T, error)
type ClusterAware[T any] func(clusterName string, cluster cluster.Cluster) T
type ClustersAware[T any] func(ctx context.Context, logger logging.Logger, clusters Clusters) (T, error)

type ClusterMatcher func(clusterId string) (clusterName string, equal bool)

type ControllerManager interface {
	GetName() string
	GetManager() mcctrl.Manager
	GetMainCluster() ClusterEquivalent
	MapTechnicalName(name string) ClusterEquivalent
	GetClusters() Clusters
	GetIndex(name string) Index
	GetIndices() Indices

	GetLogger() logging.Logger
	GetControllerDefinition(name string) ControllerDefinition
}

type ClusterNames = set.Set[string]

type ControllerDefinition interface {
	flagutils.Options
	GetName() string

	GetActivationConstraints() constraints.Constraints
	GetCluster() string
	GetClusters() ClusterNames
	GetResource() client.Object
	GetGroups() set.Set[string]
	GetWatchPredicates() []predicate.Predicate

	GetRequiredClusters(mappings ControllerMappings) ClusterNames
	GetError() error
	GetOptions() flagutils.Options

	// CreateIndices creates and exports locally defined indices prior to controller creation.
	CreateIndices(ctx context.Context, mapping ControllerMappings, mgr ControllerManager) error

	// CreateController handles the global definitions and provides
	// a Controller
	CreateController(ctx context.Context, mapping ControllerMappings, mgr ControllerManager) (Controller, error)
}

type ControllerNames = constraints.ControllerNames

type MappedControllerDefinition interface {
	ControllerDefinition
}

type Controller interface {
	GetName() string
	GetOptions() flagutils.Options
	GetFieldManager() string
	GetLogger() logging.Logger
	GetClusterMappings() Mappings
	GetClusters() Clusters
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

type Cluster interface {
	ClusterEquivalent

	client.Client
	cluster.Cluster
	// Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
	// List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error

	WaitForCacheSync(context.Context) bool

	Unwrap() Cluster
	GetCluster() cluster.Cluster

	GetIndex(name string) Index

	GetTypeConverter() merge.Converters

	IsSameAs(ClusterEquivalent) bool

	GetAPIServerURL() (*url.URL, error)
}

type Clusters interface {
	internal.Group[ClusterEquivalent]
	IsMulti() bool
	GetClusterById(clusterId string) ClusterEquivalent
}

type Index interface {
	GetName() string
	GetCluster() ClusterEquivalent
	GetGVK() schema.GroupVersionKind

	GetList(ctx context.Context, namespace, key string) (client.ObjectList, error)

	ForEachItem(ctx context.Context, namespace, key string, action func(ctx context.Context, obj runtime.Object) error) error
	Trigger(ctx context.Context, namespace, key string) error

	GetEffective() Index
	GetResource() client.Object
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

	RequireIdentity()

	GetConfig(*config.ConfigOptions) (*config.Config, error)

	GetName() string
	GetFallback() string
	GetDescription() string
	GetScheme() *runtime.Scheme
	AcceptFleet() bool

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

	GetType() FleetType

	GetProvider() multicluster.Provider
	GetClusterNames() []string
	GetCluster(name string) Cluster
	GetClusterByLocalName(name string) Cluster

	GetClusterById(id string) ClusterEquivalent

	GetBaseCluster() Cluster

	Compose(name string) string
}

type FleetType interface {
	GetType() string
	Create(defs ClusterDefinitions, definition ClusterDefinition, config config.Config, log logging.Logger) (Fleet, error)
	GetRules(def ClusterDefinition) config.Rules
}

type ClusterEquivalent interface {
	client.FieldIndexer
	enqueue.TypedMux[mcreconcile.Request]

	GetName() string
	GetId() string
	GetInfo() string
	GetTypeInfo() string

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

	// ListGlobalKeys provides a list of global object keys for objects
	// for the given object type.
	//
	// For fleets the complete cluster set represented by this cluster equivalent is used.
	ListGlobalKeys(ctx context.Context, obj runtime.Object, opts ...client.ListOption) ([]GlobalKey, error)

	// ListIndexedGlobalKeysByObjectKey provides a list of global object keys for objects indexed with the given index and key
	// for the given object type.
	//
	// For fleets the index refers to the complete cluster set represented by this cluster equivalent.
	// The list options should not contain an index specification.
	// The key is again a global object key. The cluster name must be specified. The group/kind information
	// is optional, depending on the used index function.
	//
	// For fleet indices the index function may return an empty cluster name for referring to the local cluster,
	// because the MCRT library does not support cluster-aware index functions. The implementation for a fleet
	// must compensate this.
	ListIndexedGlobalKeysByObjectKey(ctx context.Context, obj runtime.Object, index string, key TypedGlobalKey, opts ...client.ListOption) ([]GlobalKey, error)

	AsCluster() Cluster
	AsFleet() Fleet

	IsSameAs(o ClusterEquivalent) bool

	Filter(clusterName string, cluster cluster.Cluster) bool
	FilterById(clusterId string) bool
	Match(clusterName string) bool
	LiftTechnical(clusterName string) (string, Cluster)
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
