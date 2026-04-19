package cluster

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/mandelsoft/goutils/funcs"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/enqueue"
	"github.com/mandelsoft/kubecrtutils/merge"
	"github.com/mandelsoft/kubecrtutils/setup"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type Syncer func()

type _cluster struct {
	lock sync.Mutex
	cluster.Cluster
	client.Client
	enqueue.TypedMux[mcreconcile.Request]

	name      string
	id        string
	converter merge.Converters
	start     sync.Once
	indices   map[string]Index
	syncer    Syncer
	sync      sync.Once
	apiServer *url.URL
}

var _ SchemeProvider = (*_cluster)(nil)

func NewClusterForCRTCluster(name string, c cluster.Cluster, opts ...any) (Cluster, error) {

	id := name
	var cl client.Client
	var conv merge.Converters
	var apiServer *url.URL
	var syncer Syncer
	for _, o := range opts {
		if o == nil {
			continue
		}
		switch v := o.(type) {
		case string:
			id = v
		case merge.Converters:
			conv = v
		case client.Client:
			cl = v
		case *url.URL:
			apiServer = v
		case Syncer:
			syncer = v
		default:
			return nil, fmt.Errorf("unknown option type %T", v)
		}
	}

	if c != nil {
		if conv == nil {
			var err error
			conv, err = merge.NewConverters(c.GetConfig())
			if err != nil {
				return nil, fmt.Errorf("cluster %v: cannot create converter: %w", funcs.First(rest.DefaultServerUrlFor(c.GetConfig())), err)
			}
		}
		cl = c.GetClient()
	}

	r := &_cluster{
		Cluster:   c,
		Client:    cl,
		name:      name,
		id:        id,
		converter: conv,
		indices:   map[string]Index{},
		apiServer: apiServer,
		syncer:    syncer,
	}
	r.TypedMux = enqueue.NewTypedMux[mcreconcile.Request](c.GetScheme(), r.createRequest)
	return r, nil
}

func NewCluster(name string, config *config.Config, opts ...cluster.Option) (Cluster, error) {
	c, err := cluster.New(config.RestConfig, opts...)
	if err != nil {
		return nil, err
	}
	id := config.GetId()
	if id == "" {
		id = name
	}
	conv, err := merge.NewConverters(config.RestConfig)
	if err != nil {
		return nil, err
	}
	r := &_cluster{
		Cluster:   c,
		Client:    c.GetClient(),
		name:      name,
		id:        id,
		converter: conv,
		indices:   map[string]Index{},
	}
	r.TypedMux = enqueue.NewTypedMux[mcreconcile.Request](c.GetScheme(), r.createRequest)
	return r, nil
}

func (c *_cluster) GetName() string {
	return c.name
}

func (c *_cluster) GetInfo() string {
	return c.GetConfig().Host
}

func (c *_cluster) GetTypeInfo() string {
	return "cluster"
}

func (c *_cluster) GetAPIServerURL() (*url.URL, error) {
	if c.apiServer != nil {
		return c.apiServer, nil
	}
	u, _, err := rest.DefaultServerUrlFor(c.GetConfig())
	return u, err
}

func (c *_cluster) Unwrap() Cluster {
	return nil
}

func (c *_cluster) GetId() string {
	return c.id
}

func (c *_cluster) GetEffective() ClusterEquivalent {
	return c
}

func (c *_cluster) AsCluster() Cluster {
	return c
}

func (c *_cluster) AsFleet() types.Fleet {
	return nil
}

func (c *_cluster) Filter(clusterName string, _ cluster.Cluster) bool {
	return c.Match(clusterName)
}

func (c *_cluster) Match(clusterName string) bool {
	return c.GetName() == Normalize(clusterName)
}

func (c *_cluster) FilterById(clusterId string) bool {
	return clusterId == c.GetId()
}

func (c *_cluster) GetClusterById(clusterId string) ClusterEquivalent {
	if clusterId == c.GetId() {
		return c
	}
	return nil
}

func (c *_cluster) LiftTechnical(clusterName string) (string, Cluster) {
	if c.name == Normalize(clusterName) {
		return c.name, c
	}
	setup.Log.Warn("problem in cluster name mappings: cluster {{cluster}} cannot be seen as {{effective}}", "effective", c.name, "cluster", clusterName)
	// is called by MCRT for omitted clusters, also
	// panic(fmt.Errorf("problem in cluster name mappings: cluster %s cannot be seen as %q", c.name, clusterName))
	return "", nil
}

func (c *_cluster) IsSameAs(o ClusterEquivalent) bool {
	if o == nil {
		return false
	}
	oc := o.AsCluster()
	if oc == nil {
		return false
	}
	return c.GetClient() == oc.GetClient()
}

func (c *_cluster) GetCluster() cluster.Cluster {
	return c.Cluster
}

func (c *_cluster) GetTypeConverter() merge.Converters {
	return c.converter
}

func (c *_cluster) createRequest(_ context.Context, key client.ObjectKey) (mcreconcile.Request, error) {
	return mcreconcile.Request{ClusterName: c.GetName(), Request: reconcile.Request{key}}, nil
}

// shitty manager API always creates an own cluster
// based on a rest config,
// which is implicitly added as runnable.
// It is not possible to pass a preconfigured cluster object.
// The workaround to fake NewClient and NewCache via options
// (see in the manager creation in package manageroptions)
// would result in calling Start on the Cache twice.
// It is not possible to ignore the call to the cache,
// because it is required to pass the synchronization barrier.
// Therefore, we have to assure that the cache start starts
// the complete cluster, but only once.
// In addition, we do not add the maon cluster object to the
// manager. This is implicitly done by providing the cache of
// this cluster.
func (c *_cluster) Start(ctx context.Context) error {
	var err error
	c.start.Do(func() {
		err = c.Cluster.Start(ctx)
	})
	return err
}

func (c *_cluster) GetCache() cache.Cache {
	return &cacheWrapper{Cache: c.Cluster.GetCache(), cluster: c}
}

type cacheWrapper struct {
	cache.Cache
	cluster *_cluster
}

// Start of the cache (provided to configure the main cluster in the
// manager, now start the complete cluster it is taken from.
// This cluster is then NOT added as runnable to the manager,
// but started via the provided cache object, which is used
// to setup the implicit cluster always created by the manager.
func (w *cacheWrapper) Start(ctx context.Context) error {
	return w.cluster.Start(ctx)
}

func (c *_cluster) WaitForCacheSync(ctx context.Context) bool {
	b := true
	c.sync.Do(func() {
		b = c.GetCache().WaitForCacheSync(ctx)
		if c.syncer != nil {
			c.syncer()
		}
	})
	return b
}

func (c *_cluster) GetIndex(name string) Index {
	return c.indices[name]
}

func (c *_cluster) IndexField(ctx context.Context, proto client.Object, name string, indexer client.IndexerFunc) error {
	return c.GetFieldIndexer().IndexField(ctx, proto, name, indexer)
}

func (c *_cluster) CreateIndex(ctx context.Context, name string, proto client.Object, indexer client.IndexerFunc, wrap ...func(cluster types.ClusterEquivalent, name string) (Index, error)) (Index, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.indices[name] != nil {
		return nil, fmt.Errorf("index %q already defined", name)
	}
	err := c.GetFieldIndexer().IndexField(ctx, proto, name, indexer)
	if err != nil {
		return nil, err
	}

	var idx Index

	w := general.Optional(wrap...)
	if w != nil {
		idx, err = w(c, name)
		if err != nil {
			return nil, err
		}
	} else {
		idx, err = cacheindex.NewDefaultIndex(name, c, proto)
		if err != nil {
			return nil, err
		}
	}

	c.indices[name] = idx
	return idx, nil
}
