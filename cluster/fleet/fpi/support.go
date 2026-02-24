package fpi

import (
	"context"
	"fmt"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/enqueue"
	"github.com/mandelsoft/kubecrtutils/setup"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

var ErrClusterNotFoundInContext = fmt.Errorf("cluster not found in context")
var ErrClusterNotOwned = fmt.Errorf("cluster not owned by fleet")

type Support struct {
	self types.Fleet
	Composer
	id      Composer
	lock    sync.Mutex
	typ     types.FleetType
	indices map[string]types.Index
	enqueue.TypedMux[mcreconcile.Request]
	scheme   *runtime.Scheme
	provider multicluster.Provider
	base     types.Cluster
}

func NewSupport(self types.Fleet, typ types.FleetType, name, id string, scheme *runtime.Scheme, base types.Cluster, provider multicluster.Provider) Support {
	if id == "" {
		id = name
	}
	s := Support{
		self:     self,
		Composer: Composer{name},
		typ:      typ,
		id:       Composer{id},
		scheme:   scheme,
		base:     base,
		provider: provider,
		indices:  make(map[string]types.Index),
	}
	s.TypedMux = enqueue.NewTypedMux[mcreconcile.Request](scheme, s.createRequest)
	return s
}

func (s *Support) Filter(clusterName string, cluster cluster.Cluster) bool {
	return s.Match(clusterName)
}

func (s *Support) Match(clusterName string) bool {
	b, _ := Split(clusterName)
	return b == s.name
}

func (s *Support) FilterById(clusterId string) bool {
	b, _ := Split(clusterId)
	return b == s.GetId()
}

func (s *Support) LiftTechnical(clusterName string) (string, types.Cluster) {
	b, n := Split(clusterName)
	if b == s.name {
		return clusterName, s.self.GetClusterByLocalName(n)
	}
	setup.Log.Error("technical cluster {{cluster}} does not match fleet {{effective}}", "cluster", clusterName, "effective", s.name)
	return "", nil
}

func (s *Support) createRequest(ctx context.Context, key client.ObjectKey) (mcreconcile.Request, error) {
	c, err := s._getCluster(ctx)
	if err != nil {
		return mcreconcile.Request{}, err
	}
	return mcreconcile.Request{ClusterName: c.GetName(), Request: reconcile.Request{key}}, nil
}

func (s *Support) _getCluster(ctx context.Context) (types.Cluster, error) {
	c := clustercontext.ClusterFor(ctx)
	if c == nil {
		return nil, ErrClusterNotFoundInContext
	}
	b, _ := Split(c.GetName())
	if b != s.name {
		return nil, ErrClusterNotOwned
	}
	return c, nil
}

func (s *Support) GetType() types.FleetType {
	return s.typ
}

func (s *Support) GetId() string {
	return s.id.GetName()
}

func (s *Support) GetBaseCluster() types.Cluster {
	return s.base
}

func (s *Support) GetScheme() *runtime.Scheme {
	return s.scheme
}

func (s *Support) GetProvider() multicluster.Provider {
	return s.provider
}

func (s *Support) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return s.GetProvider().IndexField(ctx, obj, field, extractValue)
}

func (s *Support) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	c, err := s._getCluster(ctx)
	if err != nil {
		return err
	}
	return c.Get(ctx, key, obj, opts...)
}

func (s *Support) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	c, err := s._getCluster(ctx)
	if err != nil {
		return err
	}
	return c.List(ctx, list, opts...)
}

func (s *Support) CreateIndex(ctx context.Context, name string, proto client.Object, indexer client.IndexerFunc, wrap ...func(cluster types.ClusterEquivalent, name string) (types.Index, error)) (types.Index, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.indices[name] != nil {
		return nil, fmt.Errorf("index %q already defined", name)
	}
	err := s.provider.IndexField(ctx, proto, name, indexer)
	if err != nil {
		return nil, err
	}

	var idx types.Index

	w := general.Optional(wrap...)
	if w != nil {
		idx, err = w(s.self, name)
		if err != nil {
			return nil, err
		}
	} else {
		idx, err = cacheindex.NewDefaultIndex(name, s.self, proto)
		if err != nil {
			return nil, err
		}
	}

	s.indices[name] = idx
	return idx, nil
}
