package kcp

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/kcp-dev/multicluster-provider/apiexport"
	"github.com/mandelsoft/goutils/maputils"
	cluster2 "github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
)

type _fleet struct {
	fpi.Support
	wrapper       wrapper
	registrations registrations
}

var _ fleet.Fleet = (*_fleet)(nil)

func New(t types.FleetType, name, id string, cfg *rest.Config, endpointSliceName string, options apiexport.Options) (*_fleet, error) {

	if id == "" {
		id = name
	}
	cl, err := cluster.New(cfg, func(opts *cluster.Options) { opts.Scheme = options.Scheme })
	if err != nil {
		return nil, err
	}
	base := cluster2.NewClusterForCRTCluster(name+"#", cl)
	conv := base.GetTypeConverter()
	p, err := apiexport.New(cfg, endpointSliceName, options)
	if err != nil {
		return nil, err
	}
	f := &_fleet{
		wrapper: wrapper{
			Provider: p,
		},
		registrations: registrations{
			Composer:  fpi.NewComposer(name),
			id:        fpi.NewComposer(id),
			converter: conv,
			clusters:  map[string]types.Cluster{},
			log:       options.Log,
		},
	}
	f.Support = fpi.NewSupport(f, t, name, id, options.Scheme, base, &f.wrapper)
	f.wrapper.registrations = &f.registrations
	return f, nil
}

func (k *_fleet) GetCluster(name string) types.Cluster {
	return k.registrations.GetCluster(name)
}

func (k *_fleet) GetClusterById(id string) types.ClusterEquivalent {
	b, n := k.Split(id)
	if b != k.GetId() {
		return nil
	}
	return k.registrations.GetCluster(k.Compose(n))
}

func (k *_fleet) GetInfo() string {
	return k.GetBaseCluster().GetConfig().Host
}

func (k *_fleet) GetClusterNames() []string {
	return k.registrations.GetClusterNames()
}

func (k *_fleet) GetEffective() types.ClusterEquivalent {
	return k
}

func (k *_fleet) AsCluster() types.Cluster {
	return nil
}

func (k *_fleet) AsFleet() types.Fleet {
	return k
}

func (k *_fleet) IsSameAs(o types.ClusterEquivalent) bool {
	if o == nil {
		return false
	}
	of := o.AsFleet()
	if of == nil {
		return false
	}
	if kcp, ok := of.(*_fleet); ok {
		return k == kcp
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

// wrapper is a Provider catching the Aware for bookkeeping of fleet clusters.
type wrapper struct {
	*apiexport.Provider
	registrations *registrations
}

func (w *wrapper) Start(ctx context.Context, aware multicluster.Aware) error {
	w.registrations.aware = aware
	return w.Provider.Start(ctx, w.registrations)
}

////////////////////////////////////////////////////////////////////////////////

type registrations struct {
	fpi.Composer
	id fpi.Composer

	lock      sync.Mutex
	converter managedfields.TypeConverter
	clusters  map[string]types.Cluster
	aware     multicluster.Aware
	log       *logr.Logger
}

func (r *registrations) GetClusterNames() []string {
	r.lock.Lock()
	defer r.lock.Unlock()
	return maputils.Keys(r.clusters)
}

func (r *registrations) GetCluster(name string) types.Cluster {
	r.lock.Lock()
	defer r.lock.Unlock()
	f, n := r.Split(name)
	if f != r.GetName() {
		return nil
	}
	return r.clusters[n]
}

func (r *registrations) Engage(ctx context.Context, name string, cluster cluster.Cluster) error {
	id := r.id.Compose(name)
	r.lock.Lock()
	r.clusters[name] = cluster2.NewClusterForCRTCluster(r.Compose(name), cluster, r.converter, id)
	r.lock.Unlock()

	r.log.Info("engage fleet cluster {{cluster}} for {{fleet}}", "cluster", name, "fleet", r.GetName())
	go func() {
		<-ctx.Done()
		r.lock.Lock()
		defer r.lock.Unlock()
		r.log.Info("disengage fleet cluster {{cluster}} for fleet {{fleet}}", "cluster", name, "fleet", r.GetName())
		delete(r.clusters, name)
	}()
	return r.aware.Engage(ctx, name, cluster)
}
