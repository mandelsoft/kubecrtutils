package replication

import (
	"context"
	"fmt"
	"sync"

	"github.com/mandelsoft/kubecrtutils/cluster/cluster"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/logging"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type ReplicationContext interface {
	context.Context
	cluster.Cluster
	logging.Logger
}

func WithCluster(ctx ReplicationContext, cl cluster.Cluster) ReplicationContext {
	return &_context{
		ctx, cl,
	}
}

type base interface {
	context.Context
	logging.Logger
}

type _context struct {
	// omit cluster view
	base
	cluster.Cluster
}

type ResourceMapping interface {
	GetGroupKind() schema.GroupKind

	GetOriginal(key client.ObjectKey) *mcreconcile.Request
	SetOriginal(ctx ReplicationContext, key client.ObjectKey, tgt mcreconcile.Request) reconcile.Problem
	RemoveOriginal(ctx ReplicationContext, key client.ObjectKey) reconcile.Problem
}

type Mapping interface {
	ForResource(kind schema.GroupKind) ResourceMapping
}

type _mapping struct {
	lock  sync.RWMutex
	types map[schema.GroupKind]*_resource
}

func NewMapping() Mapping {
	return &_mapping{types: make(map[schema.GroupKind]*_resource)}
}

func (m *_mapping) ForResource(kind schema.GroupKind) ResourceMapping {
	m.lock.RLock()
	r := m.types[kind]
	if r != nil {
		return r
	}
	m.lock.RUnlock()

	m.lock.Lock()
	defer m.lock.Unlock()
	r = m.types[kind]
	if r != nil {
		return r
	}
	r = &_resource{gk: kind, index: newResourceIndex(), base: m}
	m.types[kind] = r
	return r
}

func (m *_mapping) assureNamespace(ctx ReplicationContext, namespace string) reconcile.Problem {
	var ns v1.Namespace
	err := ctx.Get(ctx, client.ObjectKey{Name: namespace}, &ns)
	if err != nil {
		if errors.IsNotFound(err) {
			ctx.Info("assure target namespace", "namespace", namespace)
			ns.Name = namespace
			err = ctx.Create(ctx, &ns)
		}
	} else {
		if ns.DeletionTimestamp != nil {
			ctx.Info("target namespace {{namespace}} already deleting -> delay until operation finished", "namespace", namespace)
			err = fmt.Errorf("target namespace already deleting")
		}
	}
	return reconcile.TemporaryProblem(err)
}

func (m *_mapping) removeNamespace(ctx ReplicationContext, namespace string) reconcile.Problem {
	ctx.Info("cleanup target namespace {{targetnamespace}}", "targetnamespace", namespace)
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	err := ctx.Delete(ctx, &ns)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.TemporaryProblem(err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type _resource struct {
	base  *_mapping
	index *_resourceIndex
	gk    schema.GroupKind
}

func (m *_resource) GetGroupKind() schema.GroupKind {
	return m.gk
}

func (m *_resource) GetOriginal(key client.ObjectKey) *mcreconcile.Request {
	m.base.lock.RLock()
	defer m.base.lock.RUnlock()
	return m.index.get(key)
}

func (m *_resource) SetOriginal(ctx ReplicationContext, key client.ObjectKey, tgt mcreconcile.Request) reconcile.Problem {
	m.base.lock.Lock()
	defer m.base.lock.Unlock()
	p := m.base.assureNamespace(ctx, key.Namespace)
	if p != nil {
		return p
	}
	m.index.add(key, tgt)
	return nil
}

func (m *_resource) RemoveOriginal(ctx ReplicationContext, key client.ObjectKey) reconcile.Problem {
	m.base.lock.Lock()
	defer m.base.lock.Unlock()

	empty := true
	for _, k := range m.base.types {
		if k != m && len(k.index.entries) != 0 {
			if len(k.index.entries[key.Namespace]) > 0 {
				empty = false
				break
			}
		}
	}

	obs := m.index.entries[key.Namespace]
	if empty && lastEntry(obs, key.Name) {
		p := m.base.removeNamespace(ctx, key.Namespace)
		if p != nil {
			return p
		}
	} else {
		s := ""
		if !empty {
			s = " and other mappings not empty"
		}
		no := len(obs)
		if no > 0 {
			if _, ok := obs[key.Name]; ok {
				no--
			}
		}
		ctx.Info("still {{no}} mappings in target namespace {{targetnamespace}}"+s, "no", no, "targetnamespace", key.Namespace)
	}
	m.index.remove(key)
	return nil
}

type _resourceIndex struct {
	entries map[string]map[string]mcreconcile.Request
}

func newResourceIndex() *_resourceIndex {
	return &_resourceIndex{
		entries: map[string]map[string]mcreconcile.Request{},
	}
}

func (i *_resourceIndex) get(key client.ObjectKey) *mcreconcile.Request {
	obs := i.entries[key.Namespace]
	if len(obs) == 0 {
		return nil
	}
	r, ok := obs[key.Name]
	if !ok {
		return nil
	}
	return &r
}

func (i *_resourceIndex) add(key client.ObjectKey, tgt mcreconcile.Request) {
	obs := i.entries[key.Namespace]
	if obs == nil {
		obs = make(map[string]mcreconcile.Request)
		i.entries[key.Namespace] = obs
	}
	obs[key.Name] = tgt
}

func (i *_resourceIndex) remove(key client.ObjectKey) {
	obs := i.entries[key.Namespace]
	if obs != nil {
		delete(obs, key.Name)
		if len(obs) == 0 {
			delete(i.entries, key.Namespace)
		}
	}
}

func lastEntry[T any](m map[string]T, key string) bool {
	if len(m) != 1 {
		return len(m) == 0
	}
	for k := range m {
		if k == key {
			return true
		}
	}
	return false
}
