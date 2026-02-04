/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler

import (
	"context"
	"time"

	"github.com/modern-go/reflect2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/priorityqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

// EventHandler enqueues reconcile.Requests in response to events (e.g. Pod Create).  EventHandlers map an Event
// for one object to trigger Reconciles for either the same object or different objects - e.g. if there is an
// Event for object with type Foo (using source.Kind) then reconcile one or more object(s) with type Bar.
//
// Identical reconcile.Requests will be batched together through the queuing mechanism before reconcile is called.
//
// * Use EnqueueRequestForObject to reconcile the object the event is for
// - do this for events for the type the Controller Reconciles. (e.g. Deployment for a Deployment Controller)
//
// * Use EnqueueRequestForOwner to reconcile the owner of the object the event is for
// - do this for events for the types the Controller creates.  (e.g. ReplicaSets created by a Deployment Controller)
//
// * Use EnqueueRequestsFromMapFunc to transform an event for an object to a reconcile of an object
// of a different type - do this for events for types the Controller may be interested in, but doesn't create.
// (e.g. If Foo responds to cluster size events, map Node events to Foo objects.)
//
// Unless you are implementing your own EventHandler, you can ignore the functions on the EventHandler interface.
// Most users shouldn't need to implement their own EventHandler.

type EventHandler = TypedEventHandler[mcreconcile.Request]

type TypedEventHandler[R comparable] interface {
	// Create is called in response to a create event - e.g. Pod Creation.
	Create(context.Context, event.CreateEvent, TypedQueueingInterface[R])

	// Update is called in response to an update event -  e.g. Pod Updated.
	Update(context.Context, event.UpdateEvent, TypedQueueingInterface[R])

	// Delete is called in response to a delete event - e.g. Pod Deleted.
	Delete(context.Context, event.DeleteEvent, TypedQueueingInterface[R])

	// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
	// external trigger request - e.g. reconcile Autoscaling, or a Webhook.
	Generic(context.Context, event.GenericEvent, TypedQueueingInterface[R])
}

var _ EventHandler = Funcs{}

// Funcs implements eventhandler.
type Funcs = TypedFuncs[mcreconcile.Request]

// TypedFuncs implements eventhandler.
type TypedFuncs[R comparable] struct {
	// Create is called in response to an add event.  Defaults to no-op.
	// RateLimitingInterface is used to enqueue reconcile.Requests.
	CreateFunc func(context.Context, event.CreateEvent, TypedQueueingInterface[R])

	// Update is called in response to an update event.  Defaults to no-op.
	// RateLimitingInterface is used to enqueue reconcile.Requests.
	UpdateFunc func(context.Context, event.UpdateEvent, TypedQueueingInterface[R])

	// Delete is called in response to a delete event.  Defaults to no-op.
	// RateLimitingInterface is used to enqueue reconcile.Requests.
	DeleteFunc func(context.Context, event.DeleteEvent, TypedQueueingInterface[R])

	// GenericFunc is called in response to a generic event.  Defaults to no-op.
	// RateLimitingInterface is used to enqueue reconcile.Requests.
	GenericFunc func(context.Context, event.GenericEvent, TypedQueueingInterface[R])
}

// Create implements EventHandler.
func (h TypedFuncs[R]) Create(ctx context.Context, e event.CreateEvent, q TypedQueueingInterface[R]) {
	if h.CreateFunc != nil {
		if !IsPriorityQueue(q) || reflect2.IsNil(e.Object) {
			h.CreateFunc(ctx, e, q)
			return
		}

		wq := workqueueWithDefaultPriority[R]{
			// We already know that we have a priority queue, that event.Object implements
			// client.Object and that its not nil
			TypedPriorityQueueingInterface: q.(TypedPriorityQueueingInterface[R]),
		}
		if e.IsInInitialList {
			wq.priority = ptr.To(LowPriority)
		}
		h.CreateFunc(ctx, e, wq)
	}
}

// Delete implements EventHandler.
func (h TypedFuncs[R]) Delete(ctx context.Context, e event.DeleteEvent, q TypedQueueingInterface[R]) {
	if h.DeleteFunc != nil {
		h.DeleteFunc(ctx, e, q)
	}
}

// Update implements EventHandler.
func (h TypedFuncs[R]) Update(ctx context.Context, e event.UpdateEvent, q TypedQueueingInterface[R]) {
	if h.UpdateFunc != nil {
		if !IsPriorityQueue(q) || reflect2.IsNil(e.ObjectOld) || reflect2.IsNil(e.ObjectNew) {
			h.UpdateFunc(ctx, e, q)
			return
		}

		wq := workqueueWithDefaultPriority[R]{
			// We already know that we have a priority queue, that event.ObjectOld and ObjectNew implement
			// client.Object and that they are  not nil
			TypedPriorityQueueingInterface: q.(TypedPriorityQueueingInterface[R]),
		}
		if any(e.ObjectOld).(client.Object).GetResourceVersion() == any(e.ObjectNew).(client.Object).GetResourceVersion() {
			wq.priority = ptr.To(LowPriority)
		}
		h.UpdateFunc(ctx, e, wq)
	}
}

// Generic implements EventHandler.
func (h TypedFuncs[R]) Generic(ctx context.Context, e event.GenericEvent, q TypedQueueingInterface[R]) {
	if h.GenericFunc != nil {
		h.GenericFunc(ctx, e, q)
	}
}

// LowPriority is the priority set by WithLowPriorityWhenUnchanged
const LowPriority = -100

// WithLowPriorityWhenUnchanged reduces the priority of events stemming from the initial listwatch or from a resync if
// and only if a priorityqueue.PriorityQueue is used. If not, it does nothing.
func WithLowPriorityWhenUnchanged(u EventHandler) EventHandler {
	// TypedFuncs already implements this so just wrap
	return Funcs{
		CreateFunc:  u.Create,
		UpdateFunc:  u.Update,
		DeleteFunc:  u.Delete,
		GenericFunc: u.Generic,
	}
}

type workqueueWithDefaultPriority[R comparable] struct {
	TypedPriorityQueueingInterface[R]
	priority *int
}

func (w workqueueWithDefaultPriority[R]) Add(item R) {
	w.TypedPriorityQueueingInterface.AddWithOpts(priorityqueue.AddOpts{Priority: w.priority}, item)
}

func (w workqueueWithDefaultPriority[R]) AddAfter(item R, after time.Duration) {
	w.TypedPriorityQueueingInterface.AddWithOpts(priorityqueue.AddOpts{Priority: w.priority, After: after}, item)
}

func (w workqueueWithDefaultPriority[R]) AddRateLimited(item R) {
	w.TypedPriorityQueueingInterface.AddWithOpts(priorityqueue.AddOpts{Priority: w.priority, RateLimited: true}, item)
}

func (w workqueueWithDefaultPriority[R]) AddWithOpts(o priorityqueue.AddOpts, item R) {
	if o.Priority == nil {
		o.Priority = w.priority
	}
	w.TypedPriorityQueueingInterface.AddWithOpts(o, item)
}
