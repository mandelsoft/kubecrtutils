package handler

import (
	"context"

	"github.com/modern-go/reflect2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/priorityqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type MapFunc = handler.TypedMapFunc[client.Object, mcreconcile.Request]

func EnqueueRequestsFromMapFunc(fn MapFunc) EventHandler {
	return &enqueueRequestsFromMapFunc{
		toRequests: fn,
	}
}

var _ EventHandler = &enqueueRequestsFromMapFunc{}

type enqueueRequestsFromMapFunc struct {
	// Mapper transforms the argument into a slice of keys to be reconciled
	toRequests MapFunc
}

// Create implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Create(
	ctx context.Context,
	evt event.CreateEvent,
	q QueueingInterface,
) {
	reqs := map[mcreconcile.Request]empty{}

	var lowPriority bool
	if IsPriorityQueue(q) && !reflect2.IsNil(evt.Object) {
		if evt.IsInInitialList {
			lowPriority = true
		}
	}
	e.mapAndEnqueue(ctx, q, evt.Object, reqs, lowPriority)
}

// Update implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Update(
	ctx context.Context,
	evt event.UpdateEvent,
	q QueueingInterface,
) {
	var lowPriority bool
	if IsPriorityQueue(q) && !reflect2.IsNil(evt.ObjectOld) && !reflect2.IsNil(evt.ObjectNew) {
		lowPriority = any(evt.ObjectOld).(client.Object).GetResourceVersion() == any(evt.ObjectNew).(client.Object).GetResourceVersion()
	}
	reqs := map[mcreconcile.Request]empty{}
	e.mapAndEnqueue(ctx, q, evt.ObjectOld, reqs, lowPriority)
	e.mapAndEnqueue(ctx, q, evt.ObjectNew, reqs, lowPriority)
}

// Delete implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Delete(
	ctx context.Context,
	evt event.DeleteEvent,
	q QueueingInterface,
) {
	reqs := map[mcreconcile.Request]empty{}
	e.mapAndEnqueue(ctx, q, evt.Object, reqs, false)
}

// Generic implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Generic(
	ctx context.Context,
	evt event.GenericEvent,
	q QueueingInterface,
) {
	reqs := map[mcreconcile.Request]empty{}
	e.mapAndEnqueue(ctx, q, evt.Object, reqs, false)
}

func (e *enqueueRequestsFromMapFunc) mapAndEnqueue(
	ctx context.Context,
	q QueueingInterface,
	o client.Object,
	reqs map[mcreconcile.Request]empty,
	lowPriority bool,
) {
	for _, req := range e.toRequests(ctx, o) {
		_, ok := reqs[req]
		if !ok {
			if lowPriority {
				q.(PriorityQueueingInterface).AddWithOpts(priorityqueue.AddOpts{
					Priority: ptr.To(handler.LowPriority),
				}, req)
			} else {
				q.Add(req)
			}
			reqs[req] = empty{}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type empty struct{}
