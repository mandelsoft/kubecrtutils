package handler

import (
	"time"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller/priorityqueue"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type QueueingInterface = TypedQueueingInterface[mcreconcile.Request]

type TypedQueueingInterface[R comparable] interface {
	Add(item R)
	AddAfter(item R, duration time.Duration)
	AddRateLimited(item R)
	NumRequeues(item R) int
}

var _ TypedQueueingInterface[mcreconcile.Request] = (workqueue.TypedRateLimitingInterface[mcreconcile.Request])(nil)
var _ TypedPriorityQueueingInterface[mcreconcile.Request] = (priorityqueue.PriorityQueue[mcreconcile.Request])(nil)

type PriorityQueueingInterface = TypedPriorityQueueingInterface[mcreconcile.Request]

type TypedPriorityQueueingInterface[R comparable] interface {
	TypedQueueingInterface[R]
	AddWithOpts(options priorityqueue.AddOpts, item ...R)
}

func IsPriorityQueue[R comparable](q TypedQueueingInterface[R]) bool {
	_, ok := q.(TypedPriorityQueueingInterface[R])
	return ok
}
