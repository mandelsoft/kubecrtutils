package builder

import (
	"context"
	"time"

	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/cluster/fleet"
	handler2 "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/priorityqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
	mchandler "sigs.k8s.io/multicluster-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func ForFleet(b *mcbuilder.Builder, fleet types.Fleet) Builder {
	return &_mcbuilder{fleet, b}
}

func isPriorityQueue[request comparable](q workqueue.TypedRateLimitingInterface[request]) bool {
	_, ok := q.(priorityqueue.PriorityQueue[request])
	return ok
}

type _mcbuilder struct {
	fleet   fleet.Fleet
	builder *mcbuilder.Builder
}

var _ Builder = (*_mcbuilder)(nil)

func (b *_mcbuilder) Named(name string) Builder {
	b.builder = b.builder.Named(name)
	return b
}

func (b *_mcbuilder) Watches(
	object client.Object,
	eventHandler handler2.EventHandler,
	opts ...WatchesOption,
) Builder {
	f := func(name string, _ cluster.Cluster) handler.TypedEventHandler[client.Object, mcreconcile.Request] {
		return &mappedHandlerF{cluster: b.fleet.GetCluster(name), handler: eventHandler}
	}

	var set watchesOptions
	for _, o := range opts {
		o.ApplyToWatches(&set)
	}
	set.applyFleetFilter(b.fleet)

	b.builder = b.builder.Watches(object, f, set.mapToMCRT()...)
	return b
}

////////////////////////////////////////////////////////////////////////////////

type mappedPrioWorkQueueF struct {
	mappedWorkQueueF
}

var _ handler2.PriorityQueueingInterface = (*mappedPrioWorkQueueF)(nil)

func (m *mappedPrioWorkQueueF) AddWithOpts(options priorityqueue.AddOpts, item ...mcreconcile.Request) {
	m.TypedRateLimitingInterface.(priorityqueue.PriorityQueue[mcreconcile.Request]).AddWithOpts(options, item...)
}

type mappedWorkQueueF struct {
	cluster string
	workqueue.TypedRateLimitingInterface[mcreconcile.Request]
}

var _ handler2.QueueingInterface = (*mappedWorkQueueF)(nil)

func (m *mappedWorkQueueF) Add(item mcreconcile.Request) {
	if item.ClusterName == "" {
		item.ClusterName = m.cluster
	}
	m.TypedRateLimitingInterface.Add(item)
}

func (m *mappedWorkQueueF) AddAfter(item mcreconcile.Request, duration time.Duration) {
	if item.ClusterName == "" {
		item.ClusterName = m.cluster
	}
	m.TypedRateLimitingInterface.AddAfter(item, duration)
}

func (m *mappedWorkQueueF) AddRateLimited(item mcreconcile.Request) {
	if item.ClusterName == "" {
		item.ClusterName = m.cluster
	}
	m.TypedRateLimitingInterface.AddRateLimited(item)
}

func (m mappedWorkQueueF) NumRequeues(item mcreconcile.Request) int {
	if item.ClusterName == "" {
		item.ClusterName = m.cluster
	}
	return m.TypedRateLimitingInterface.NumRequeues(item)
}

func mapQueue(clusterName string, w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) handler2.QueueingInterface {
	m := mappedWorkQueueF{cluster: clusterName, TypedRateLimitingInterface: w}
	if isPriorityQueue(w) {
		return &mappedPrioWorkQueueF{m}
	} else {
		return &m
	}
}

////////////////////////////////////////////////////////////////////////////////

type mappedHandlerF struct {
	cluster types.Cluster
	handler handler2.EventHandler
}

var _ mchandler.EventHandler = (*mappedHandlerF)(nil)

func (m *mappedHandlerF) _cluster(ctx context.Context) context.Context {
	return clustercontext.WithCluster(ctx, m.cluster)
}

func (m *mappedHandlerF) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	m.handler.Create(m._cluster(ctx), e, mapQueue(m.cluster.GetName(), w))
}

func (m *mappedHandlerF) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	m.handler.Update(m._cluster(ctx), e, mapQueue(m.cluster.GetName(), w))
}

func (m *mappedHandlerF) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	m.handler.Delete(m._cluster(ctx), e, mapQueue(m.cluster.GetName(), w))
}

func (m *mappedHandlerF) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	m.handler.Generic(m._cluster(ctx), e, mapQueue(m.cluster.GetName(), w))
}
