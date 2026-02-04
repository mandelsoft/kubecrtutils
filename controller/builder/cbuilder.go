package builder

import (
	"context"
	"time"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	myhandler "github.com/mandelsoft/kubecrtutils/controller/handler"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/priorityqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

func ForCluster(b *builder.Builder, cl types.Cluster) Builder {
	return &_cbuilder{cl, b}
}

type _cbuilder struct {
	cluster types.Cluster
	builder *builder.Builder
}

var _ Builder = (*_cbuilder)(nil)

func (b *_cbuilder) Named(name string) Builder {
	b.builder = b.builder.Named(name)
	return b
}

func (b *_cbuilder) Watches(
	object client.Object,
	eventHandler myhandler.EventHandler,
	opts ...WatchesOption,
) Builder {
	f := &mappedHandlerC{cluster: b.cluster, handler: eventHandler}

	var set watchesOptions
	for _, o := range opts {
		o.ApplyToWatches(&set)
	}
	b.builder = b.builder.Watches(object, f, set.mapToCRT()...)
	return b
}

////////////////////////////////////////////////////////////////////////////////

type mappedHandlerC struct {
	cluster types.Cluster
	handler myhandler.EventHandler
}

var _ handler.EventHandler = (*mappedHandlerC)(nil)

func (m *mappedHandlerC) Create(ctx context.Context, e event.CreateEvent, w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	m.handler.Create(m.withCluster(ctx), e, m.mapQueue(w))
}

func (m *mappedHandlerC) Update(ctx context.Context, e event.UpdateEvent, w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	m.handler.Update(m.withCluster(ctx), e, m.mapQueue(w))
}

func (m *mappedHandlerC) Delete(ctx context.Context, e event.DeleteEvent, w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	m.handler.Delete(m.withCluster(ctx), e, m.mapQueue(w))
}

func (m *mappedHandlerC) withCluster(ctx context.Context) context.Context {
	return clustercontext.WithCluster(ctx, m.cluster)
}

func (h *mappedHandlerC) mapQueue(w workqueue.TypedRateLimitingInterface[reconcile.Request]) myhandler.QueueingInterface {
	m := mappedWorkQueueC{cluster: h.cluster.GetName(), TypedRateLimitingInterface: w}
	if isPriorityQueue(w) {
		return &mappedPrioWorkQueueC{m}
	} else {
		return &m
	}
}

////////////////////////////////////////////////////////////////////////////////

func (m *mappedHandlerC) Generic(ctx context.Context, e event.GenericEvent, w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	m.handler.Generic(m.withCluster(ctx), e, m.mapQueue(w))
}

type mappedPrioWorkQueueC struct {
	mappedWorkQueueC
}

var _ myhandler.PriorityQueueingInterface = (*mappedPrioWorkQueueC)(nil)

func (m *mappedPrioWorkQueueC) AddWithOpts(options priorityqueue.AddOpts, item ...mcreconcile.Request) {
	m.TypedRateLimitingInterface.(priorityqueue.PriorityQueue[reconcile.Request]).AddWithOpts(options, sliceutils.Transform(item, func(item mcreconcile.Request) reconcile.Request { return item.Request })...)
}

type mappedWorkQueueC struct {
	cluster string
	workqueue.TypedRateLimitingInterface[reconcile.Request]
}

var _ myhandler.QueueingInterface = (*mappedWorkQueueC)(nil)

func (m *mappedWorkQueueC) Add(item mcreconcile.Request) {
	m.TypedRateLimitingInterface.Add(item.Request)
}

func (m *mappedWorkQueueC) AddAfter(item mcreconcile.Request, duration time.Duration) {
	m.TypedRateLimitingInterface.AddAfter(item.Request, duration)
}

func (m *mappedWorkQueueC) AddRateLimited(item mcreconcile.Request) {
	m.TypedRateLimitingInterface.AddRateLimited(item.Request)
}

func (m *mappedWorkQueueC) NumRequeues(item mcreconcile.Request) int {
	return m.TypedRateLimitingInterface.NumRequeues(item.Request)
}
