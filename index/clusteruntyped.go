package index

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterUntypedIndex interface {
	Replace(r string, clusterName string, a client.ObjectKey, b ...client.ObjectKey) bool
	Add(r string, clusterName string, a, b client.ObjectKey) bool
	Remove(r string, clusterName string, a, b client.ObjectKey) bool
	RemoveObject(clusterName string, o client.ObjectKey) bool
	UsersFor(r string, clusterName string, b client.ObjectKey) sets.Set[client.ObjectKey]
	UsedBy(r string, clusterName string, a client.ObjectKey) sets.Set[client.ObjectKey]
	Clear(r string, clusterName string) bool
}

type clusteruntyped struct {
	lock     sync.Mutex
	clusters map[string]UntypedIndex
}

func NewClusterUntypedIndex() ClusterUntypedIndex {
	return &clusteruntyped{
		clusters: make(map[string]UntypedIndex),
	}
}

func (u *clusteruntyped) Replace(r string, clusterName string, a client.ObjectKey, b ...client.ObjectKey) bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m != nil {
		return m.Replace(r, a, b...)
	}
	return false
}

func (u *clusteruntyped) Add(r string, clusterName string, a, b client.ObjectKey) bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m == nil {
		m = NewUntyped()
		u.clusters[clusterName] = m
	}
	return m.Add(r, a, b)
}

func (u *clusteruntyped) Remove(r string, clusterName string, a, b client.ObjectKey) bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m == nil {
		return false
	}
	return m.Remove(r, a, b)
}

func (u *clusteruntyped) RemoveObject(clusterName string, o client.ObjectKey) bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m == nil {
		return false
	}
	return m.RemoveObject(o)
}

func (u *clusteruntyped) UsersFor(r string, clusterName string, b client.ObjectKey) sets.Set[client.ObjectKey] {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m == nil {
		return nil
	}
	return m.UsersFor(r, b)
}

func (u *clusteruntyped) UsedBy(r string, clusterName string, a client.ObjectKey) sets.Set[client.ObjectKey] {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m == nil {
		return nil
	}
	return m.UsedBy(r, a)
}

func (u *clusteruntyped) Clear(r string, clusterName string) bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	m := u.clusters[clusterName]
	if m == nil {
		return false
	}
	b := m.Clear(r)
	if m.IsEmpty() {
		delete(u.clusters, clusterName)
	}
	return b
}
