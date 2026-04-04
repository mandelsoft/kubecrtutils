package types

import (
	"context"
	"net/url"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/enqueue"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/merge"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type ClusterNames = plain.ClusterNames

type ClusterDefinitionProvider interface {
	GetDefinition() ClusterDefinition
}

type ClusterFilter interface {
	GetUsedClusters(ConstraintContext) ClusterNames
}

type ClusterDefinition interface {
	ClusterDefinitionProvider
	flagutils.Options
	flagutils.OptionSetProvider

	RequireIdentity()

	GetConfig(options *config.ConfigOptions) (*config.Config, error)

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
	IsDisabled(name string) bool
}
